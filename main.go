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

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message},
	})
}

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
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "could not list subscriptions")
		return
	}

	OK(c, webhooks)
}
