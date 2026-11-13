package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"webhook-delivery-backend/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(os.Stderr.Name(), "Unable to create connection pool: %v\n", err)
	}
	defer pool.Close()

	queries := db.New(pool)
	handler := &WebhookHandler{queries: queries}

	router := gin.Default()
	router.GET("/subscriptions", handler.List)

	router.Run() // Listens on 0.0.0.0:8080
}

type WebhookHandler struct {
	queries *db.Queries
}

func (h *WebhookHandler) List(c *gin.Context) {
	webhooks, err := h.queries.ListWebhooks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not list subscriptions",
		})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}
