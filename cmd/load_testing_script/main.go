package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/internal/model"
)

func main() {
	cfg := config.MustLoadConfig()
	host := cfg.HTTPServer.Host
	port := cfg.HTTPServer.Port

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	for i := 0; i < 1000; i++ {
		resp, err := client.Get(fmt.Sprintf("http://%s:%d/api/order/%d", host, port, i))
		if err != nil {
			log.Fatalf("Failed to get order %d: %v", i, err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			log.Printf("Failed to read body for order %d: %v", i, err)
			continue
		}

		var response struct {
			Status string      `json:"status"`
			Data   model.Order `json:"data"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			log.Printf("Failed to unmarshal order %d: %v", i, err)
			continue
		}

		// печатаем ответ
		fmt.Printf("Order %d: status=%s, order=%+v\n", i, response.Status, response.Data)
	}
}
