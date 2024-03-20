package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/redis/go-redis/v9"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

type ConnectionInfo struct {
	Userame             string `json:"username"`
	Password            string `json:"password"`
	Engine              string `json:"engine"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Ssl                 bool   `json:"ssl"`
	DBClusterIdentifier string `json:"dbClusterIdentifier"`
}

func main() {
	port := ":80"
	http.HandleFunc("/", index)
	http.HandleFunc("/redis", cache)
	http.HandleFunc("/healthz", healthz)
	log.Println("Listen on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	secretName := os.Getenv("SECRET_NAME")
	region := os.Getenv("AWS_REGION")

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())
	}

	var secretString string = *result.SecretString

	var connectionInfo ConnectionInfo
	json.Unmarshal([]byte(secretString), &connectionInfo)

	var connectionStringTemplate = "mongodb://%s:%s@%s/?tls=true&replicaSet=rs0&readpreference=secondaryPreferred"
	connectionURI := fmt.Sprintf(connectionStringTemplate, connectionInfo.Userame, connectionInfo.Password, connectionInfo.Host)

	tlsConfig, err := getCustomTLSConfig("global-bundle.pem")
	if err != nil {
		log.Fatalf("Failed getting TLS configuration: %v", err)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to cluster: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping cluster: %v", err)
	}

	fmt.Fprintf(w, "Authentication succeed")
}

func cache(w http.ResponseWriter, r *http.Request) {
	redisAddr := os.Getenv("REDIS_ENDPOINT")
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	state, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "cached")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.New("Failed parsing pem file")
	}

	return tlsConfig, nil
}
