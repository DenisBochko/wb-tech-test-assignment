package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/internal/model"
	"wb-tech-test-assignment/pkg/kafka"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoadConfig()

	producer, err := kafka.NewProducer(
		cfg.Brokers,
		cfg.Producer.OrdersProducer.Topic,
		kafka.WithBalancer(kafka.RoundRobin),
		kafka.WithRequiredAcks(kafka.RequireAll),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = producer.Close()
	}()

	wg := &sync.WaitGroup{}
	wg.Add(cfg.Producer.WorkerCount)

	for i := 1; i <= cfg.Producer.WorkerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			worker(ctx, workerID, producer)
		}(i)
	}

	wg.Wait()
}

func worker(ctx context.Context, id int, producer kafka.Producer) {
	testOrder := model.Order{
		OrderUID:    "",
		TrackNumber: "WB123456789",
		Entry:       "WEB",

		Delivery: model.Delivery{
			Name:    "Иван Иванов",
			Phone:   "+79991234567",
			Zip:     "101000",
			City:    "Москва",
			Address: "ул. Арбат, д. 10, кв. 5",
			Region:  "Москва",
			Email:   "ivan@example.com",
		},

		Payment: model.Payment{
			Transaction:  "PAY123456",
			RequestID:    "REQ7890",
			Currency:     "RUB",
			Provider:     "bank_card",
			Amount:       3500,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Tinkoff",
			DeliveryCost: 300,
			GoodsTotal:   3200,
			CustomFee:    0,
		},

		Items: []model.Item{
			{
				ChrtID:      111111,
				TrackNumber: "WB123456789",
				Price:       2000,
				RID:         "RID12345",
				Name:        "Футболка мужская",
				Sale:        10,
				Size:        "L",
				TotalPrice:  1800,
				NmID:        555555,
				Brand:       "Nike",
				Status:      1,
			},
			{
				ChrtID:      222222,
				TrackNumber: "WB123456789",
				Price:       1200,
				RID:         "RID67890",
				Name:        "Кепка",
				Sale:        0,
				Size:        "M",
				TotalPrice:  1200,
				NmID:        666666,
				Brand:       "Adidas",
				Status:      1,
			},
		},

		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        "cust-001",
		DeliveryService:   "cdek",
		ShardKey:          "1",
		SmID:              42,
		DateCreated:       time.Now(),
		OofShard:          "2",
	}

	for i := (id - 1) * 100; i < id*100; i++ {
		testOrder.OrderUID = strconv.Itoa(i)

		massage, err := json.Marshal(testOrder)
		if err != nil {
			log.Println(err)
		}

		partition, offset, err := producer.PushMessage(ctx, []byte{byte(i)}, massage)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Pushed message to partition %d at offset %d\n", partition, offset)
	}
}
