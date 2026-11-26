package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
	"webhook-delivery-backend/internal/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Unable to load .env: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	handler := &WebhookHandler{queries: db.New(pool), pool: pool}

	// background worker
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	go Worker(workerCtx, pool, db.New(pool))

	// rest api
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/webhooks", handler.List)
	router.POST("/webhooks", handler.Create)
	router.POST("/trigger", handler.HandleEvent)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func Worker(context context.Context, pool *pgxpool.Pool, queries *db.Queries) {
	var wg sync.WaitGroup

	for {
		time.Sleep(4 * time.Second)

		if context.Err() != nil {
			break
		}

		// fetch deliveries
		pendingDeliveries, err := queries.ListPendingDeliveriesWithWebhooks(context)

		log.Printf("Worker::pendingDeliveries %v\n", pendingDeliveries)

		if err != nil {
			continue
		}

		for _, delivery := range pendingDeliveries {
			wg.Go(func() {
				queries.UpdateDeliveryStatus(context, db.UpdateDeliveryStatusParams{
					ID:     delivery.DeliveryID,
					Status: "in_progress",
				})

				log.Printf("delivery %v is in_progress\n", delivery.DeliveryID)

				err := sendWebhook(context, delivery.TargetUrl, delivery.EventData, delivery.DeliveryID)

				// send webhook failed
				if err != nil {
					queries.UpdateDeliveryStatus(context, db.UpdateDeliveryStatusParams{
						ID:     delivery.DeliveryID,
						Status: "failed",
					})
					log.Printf("delivery %v is failed\n", delivery.DeliveryID)
				} else {
					queries.UpdateDeliveryStatus(context, db.UpdateDeliveryStatusParams{
						ID:     delivery.DeliveryID,
						Status: "done",
					})
					log.Printf("delivery %v is done\n", delivery.DeliveryID)
				}

			})
		}

	}

	wg.Wait()
}
