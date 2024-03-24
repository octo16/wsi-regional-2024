package com.example.demo;

import java.time.LocalTime;
import java.time.LocalDate;
import java.time.ZoneId;
import java.util.HashMap;
import java.util.Map;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api")
@SpringBootApplication
public class DemoApplication {

	public static void main(String[] args) {
		SpringApplication.run(DemoApplication.class, args);
	}
	
	@GetMapping("/healthz")
	public String healthz() {
		return "ok";
	}
	
	@GetMapping("/user")
	public Map<String,Object> user(@RequestParam(value = "name") String name) {
		Map<String, Object> response = new HashMap<>();
		response.put("message", String.format("Hello, %s!", name));
		response.put("timestamp", LocalDate.now(ZoneId.of("GMT+09:00")) + " " + LocalTime.now(ZoneId.of("GMT+09:00")));
		return response;
	}
	
	@PostMapping("/payment")
	public Map<String,Object> payment(@RequestBody Map<String, Object> body) {
		body.put("timestamp", LocalDate.now(ZoneId.of("GMT+09:00")) + " " + LocalTime.now(ZoneId.of("GMT+09:00")));
		return body;
	}

}
