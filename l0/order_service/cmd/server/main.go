package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/kafka"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/internal/telemetry"
)

func main() {
	// telemetry
	cleanup := telemetry.InitOTel("order-service")
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// config
	cfg := config.Load()

	// postgre
	repo, err := repository.NewPostgresOrderRepo(cfg.DB)
	if err != nil {
		log.Fatal("Failed to connect to postgres:", err)
	}
	defer repo.Close()
	log.Println("Connected to PostgreSQL")

	//redis
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to redis:", err)
	}
	log.Println("Connected to Redis")
	defer redisCache.Close()

	// service
	orderService := service.NewOrderService(repo, redisCache)

	// restore fom cache
	if err := orderService.RestoreCache(ctx); err != nil {
		log.Println("Cache restore warning:", err)
	} else {
		log.Println("Cache restored from database")
	}

	// kafka consumer
	consumer := kafka.NewConsumer(cfg.Kafka, orderService)
	go consumer.Start(ctx)
	defer consumer.Stop()

	// web handler
	webHandler, err := handler.NewWebHandler()
	if err != nil {
		log.Fatal("Failed to create web handler:", err)
	}

	// order handler
	orderHandler := handler.NewOrderHandler(orderService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(otelgin.Middleware("order-service"))
	r.GET("/", webHandler.Index)
	r.GET("/static/*file", webHandler.Static)
	r.GET("/order/:id", orderHandler.GetOrder)
	r.GET("/health", orderHandler.HealthCheck)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server started on http://localhost:%s", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server failed", err)
		}
	}()

	//graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("shutting down...")
	cancel()

	shtdCtx, shtdCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shtdCancel()

	if err := srv.Shutdown(shtdCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	//time for consumer
	time.Sleep(2 * time.Second)

	log.Println("End")
}
