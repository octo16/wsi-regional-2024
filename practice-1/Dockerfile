FROM golang:alpine

WORKDIR /app
COPY global-bundle.pem .
COPY *.go ./
RUN go build -o app main.go
EXPOSE 80
CMD [ "./app" ]