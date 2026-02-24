package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/kafka"
	"order-service/internal/repository"
	"order-service/internal/service"
)

func main() {
	cfg := config.Load()

	repo, err := repository.NewPostgresOrderRepo(cfg.DB)
	if err != nil {
		log.Fatal("Failed to connect to postgres:", err)
	}
	defer repo.Close()
	log.Println("Connected to PostgreSQL")

	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to redis:", err)
	}
	log.Println("Connected to Redis")
	defer redisCache.Close()

	orderService := service.NewOrderService(repo, redisCache)

	if err := orderService.RestoreCache(); err != nil {
		log.Println("Cache restore warning:", err)
	} else {
		log.Println("Cache restored from database")
	}

	consumer := kafka.NewConsumer(cfg.Kafka, orderService)
	go consumer.Start()
	defer consumer.Stop()

	webHandler, err := handler.NewWebHandler()
	if err != nil {
		log.Fatal("Failed to create web handler:", err)
	}

	orderHandler := handler.NewOrderHandler(orderService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", webHandler.Index)
	r.GET("/static/*file", webHandler.Static)
	r.GET("/order/:id", orderHandler.GetOrder)
	r.GET("/health", orderHandler.HealthCheck)

	go func() {
		log.Printf("Server started on http://localhost:%s", cfg.HTTP.Port)
		if err := r.Run(":" + cfg.HTTP.Port); err != nil {
			log.Fatal("Server failed", err)
		}
	}()
	//graceful норм сделать
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
	time.Sleep(2 * time.Second)
	log.Println("End")
}
