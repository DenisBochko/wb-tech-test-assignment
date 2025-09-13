package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"

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
	for i := (id - 1) * 100; i < id*100; i++ {
		order := generateOrder(i)

		massage, err := json.Marshal(order)
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

func generateOrder(id int) model.Order {
	return model.Order{
		OrderUID:    strconv.Itoa(id),
		TrackNumber: gofakeit.Regex(`[A-Z]{2}[0-9]{9}`),
		Entry:       "WEB",

		Delivery: model.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Street,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},

		Payment: model.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    gofakeit.UUID(),
			Currency:     gofakeit.CurrencyShort(),
			Provider:     gofakeit.RandomString([]string{"bank_card", "paypal", "qiwi", "apple_pay"}),
			Amount:       int(gofakeit.Price(1000, 10000)),
			PaymentDt:    time.Now().Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: int(gofakeit.Price(200, 500)),
			GoodsTotal:   int(gofakeit.Price(500, 9500)),
			CustomFee:    int(gofakeit.Price(0, 300)),
		},

		Items: []model.Item{
			{
				ChrtID:      gofakeit.Number(100000, 999999),
				TrackNumber: gofakeit.Regex(`[A-Z]{2}[0-9]{9}`),
				Price:       int(gofakeit.Price(500, 5000)),
				RID:         gofakeit.UUID(),
				Name:        gofakeit.ProductName(),
				Sale:        gofakeit.Number(0, 50),
				Size:        gofakeit.RandomString([]string{"S", "M", "L", "XL"}),
				TotalPrice:  int(gofakeit.Price(400, 4500)),
				NmID:        gofakeit.Number(100000, 999999),
				Brand:       gofakeit.Car().Brand,
				Status:      1,
			},
		},

		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.RandomString([]string{"cdek", "dhl", "ups", "dpd"}),
		ShardKey:          strconv.Itoa(gofakeit.Number(1, 5)),
		SmID:              gofakeit.Number(1, 100),
		DateCreated:       time.Now(),
		OofShard:          strconv.Itoa(gofakeit.Number(1, 5)),
	}
}
