package app

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"wb-tech-test-assignment/internal/api/http/handler"
	"wb-tech-test-assignment/internal/api/http/middleware"
	"wb-tech-test-assignment/internal/config"
	"wb-tech-test-assignment/internal/repository"
	"wb-tech-test-assignment/internal/service"
	"wb-tech-test-assignment/pkg/kafka"
	"wb-tech-test-assignment/pkg/postgres"
	"wb-tech-test-assignment/pkg/server"
)

type App struct {
	Cfg        *config.Config
	Log        *zap.Logger
	DB         postgres.Postgres
	Consumer   kafka.ConsumerGroupRunner
	HTTPServer server.HTTPServer
	Service    *Service
}

type Repository struct {
	OrderRepository *repository.OrderRepository
}

type Service struct {
	OrderService *service.OrderService
}

func New(ctx context.Context, cfg *config.Config, log *zap.Logger) (*App, error) {
	db, err := initDB(&cfg.Database)
	if err != nil {
		log.Error("Failed to initialize database", zap.Error(err))

		return nil, err
	}

	consumer, err := initKafka(&cfg.Kafka, log)
	if err != nil {
		log.Error("Failed to initialize kafka", zap.Error(err))

		return nil, err
	}

	repo := initRepository(db)

	svc := initService(log, &cfg.Subscriber, consumer, db, repo)

	httpServer := initHTTPServer(ctx, log, cfg.HTTPServer, svc.OrderService)

	return &App{
		Cfg:        cfg,
		Log:        log,
		DB:         db,
		Consumer:   consumer,
		HTTPServer: httpServer,
		Service:    svc,
	}, nil
}

func MustNew(ctx context.Context, cfg *config.Config, log *zap.Logger) *App {
	app, err := New(ctx, cfg, log)
	if err != nil {
		panic(err)
	}

	return app
}

func (a *App) Run(ctx context.Context) error {
	errors := make(chan error)
	defer close(errors)

	go func() {
		if err := a.Service.OrderService.Run(ctx); err != nil {
			errors <- err
		}
	}()

	go func() {
		if err := a.HTTPServer.Run(); err != nil {
			errors <- err
		}
	}()

	if err := <-errors; err != nil {
		return err
	}

	return nil
}

func (a *App) Shutdown() error {
	a.DB.Close()

	a.Log.Debug("Database closed")

	if err := a.Service.OrderService.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown order service: %w", err)
	}

	a.Log.Debug("Order service shutdown")

	if err := a.HTTPServer.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown http server: %w", err)
	}

	a.Log.Debug("Http server shutdown")

	return nil
}

func initDB(cfg *config.Database) (postgres.Postgres, error) {
	postgresCfg := &postgres.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.Name,
		SSLMode:  cfg.SSLMode,
		MaxConns: cfg.MaxConns,
		MinConns: cfg.MinConns,
		Migration: postgres.Migration{
			Path:      cfg.Migration.Path,
			AutoApply: cfg.Migration.AutoApply,
		},
	}

	db, err := postgres.New(postgresCfg)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initKafka(cfg *config.Kafka, log *zap.Logger) (kafka.ConsumerGroupRunner, error) {
	consumerGroup, err := kafka.NewConsumerGroupRunner(
		cfg.Brokers,
		cfg.Subscriber.OrdersSubscriber.GroupID,
		[]string{cfg.Subscriber.OrdersSubscriber.Topic},
		cfg.Subscriber.OrdersSubscriber.BufferSize,
		kafka.WithBalancerConsumer(kafka.RoundrobinBalanceStrategy),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		startAndRunningStr := <-consumerGroup.Info()

		log.Info(startAndRunningStr)
	}()

	return consumerGroup, nil
}

func initRepository(db postgres.Postgres) *Repository {
	orderRepository := repository.NewOrderRepository(db.Pool())

	return &Repository{
		OrderRepository: orderRepository,
	}
}

func initService(log *zap.Logger, cfg *config.Subscriber, consumer kafka.ConsumerGroupRunner, db postgres.Postgres, repo *Repository) *Service {
	orderService := service.NewOrderService(log, cfg, consumer, db, repo.OrderRepository)

	return &Service{
		OrderService: orderService,
	}
}

func initHTTPServer(ctx context.Context, log *zap.Logger, cfg config.HTTPServer, svc *service.OrderService) server.HTTPServer {
	r := chi.NewRouter()

	r.Use(middleware.Logger(log))

	r.Get("/", handler.MainPage)
	r.Get("/api/ping", handler.Ping)
	r.Get("/api/order/{orderUID}", handler.GetOrder(ctx, svc))

	httpServer := server.NewHTTPServer(
		server.WithAddr(cfg.Host, cfg.Port),
		server.WithTimeout(cfg.Timeout.Read, cfg.Timeout.Write, cfg.Timeout.Idle),
		server.WithHandler(r),
	)

	return httpServer
}
