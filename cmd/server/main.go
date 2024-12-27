// cmd/server/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	router "github.com/iamrajjoshi/pinguin/internal/api"
	"github.com/iamrajjoshi/pinguin/internal/check"
	"github.com/iamrajjoshi/pinguin/internal/monitor"
	"github.com/iamrajjoshi/pinguin/internal/worker"

	"github.com/iamrajjoshi/pinguin/internal/scheduler"
	"github.com/iamrajjoshi/pinguin/internal/store"
)

func main() {

	// Initialize DB connection
	db, err := store.NewDB(store.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		os.Exit(1)
	}

	monitorService := monitor.NewMonitorService(db)
	checkService := check.NewCheckService(db)

	// Start Echo server
	e := router.New(db, monitorService, checkService)

	go func() {
		if err := e.Start(":8080"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	// Create and start scheduler
	scheduler := scheduler.NewScheduler(rdb, monitorService, checkService)
	scheduler.ScheduleOnStartup(context.Background())
	go func() {
		if err := scheduler.Run(context.Background()); err != nil {
			log.Printf("Scheduler error: %v", err)
			// TODO: restart scheduler
		}
	}()

	// Start workers
	for i := 0; i < 3; i++ {
		worker := worker.NewWorker(i, rdb, monitorService, checkService)
		go func(id int) {
			if err := worker.Run(context.Background()); err != nil {
				log.Printf("Worker %d error: %v", id, err)
				// TODO: restart worker, probably need a worker pool
			}
		}(i)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
