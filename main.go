package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"webhook-delivery-backend/internal/db"

	"github.com/gin-contrib/cors"
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

	queries := db.New(pool)
	handler := &WebhookHandler{queries: queries}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/webhooks", handler.List)
	router.POST("/webhooks", handler.Create)
	router.POST("/trigger", handler.HandleEvent)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
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

type CreateWebhookRequest struct {
	EventName string `json:"event_name" binding:"required"`
	TargetURL string `json:"target_url" binding:"required"`
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req CreateWebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "STATUS_BAD_REQUEST", "event_name and target_url are required")
		return
	}

	webhook, err := h.queries.CreateWebhook(c.Request.Context(), db.CreateWebhookParams{
		EventName: req.EventName,
		TargetUrl: req.TargetURL,
	})

	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "could not create webhook")
	}

	OK(c, webhook)
}

type HandleEventRequest struct {
	EventName string `json:"event_name" binding:"required"`
	Data      any    `json:"data" binding:"required"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func sendWebhook(ctx context.Context, h *WebhookHandler, targetURL string, payload any, deliveryId int64) error {
	body, err := json.Marshal(payload)
	if err != nil {
		// update status of delivery
		h.queries.UpdateDeliveryStatus(ctx, db.UpdateDeliveryStatusParams{
			ID:     deliveryId,
			Status: "failed", // TODO:
		})
		return fmt.Errorf("encode payload %w\n", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		targetURL,
		bytes.NewReader(body),
	)

	if err != nil {
		// update status of delivery
		h.queries.UpdateDeliveryStatus(ctx, db.UpdateDeliveryStatusParams{
			ID:     deliveryId,
			Status: "failed", // TODO:
		})
		return fmt.Errorf("create request %v\n", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		// update status of delivery
		h.queries.UpdateDeliveryStatus(ctx, db.UpdateDeliveryStatusParams{
			ID:     deliveryId,
			Status: "failed", // TODO:
		})
		return fmt.Errorf("send request: %v\n", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {

		// update status of delivery
		h.queries.UpdateDeliveryStatus(ctx, db.UpdateDeliveryStatusParams{
			ID:     deliveryId,
			Status: "failed", // TODO:
		})
		return fmt.Errorf("target returned HTTP %d\n", res.StatusCode)
	}

	log.Printf("sendWebhook success, updating delivery %v\n", deliveryId)

	// update status of delivery
	h.queries.UpdateDeliveryStatus(ctx, db.UpdateDeliveryStatusParams{
		ID:     deliveryId,
		Status: "done", // TODO:
	})

	return nil
}

func (h *WebhookHandler) HandleEvent(c *gin.Context) {
	var req HandleEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "STATUS_BAD_REQUEST", "event_name and data are required")
		return
	}

	// one event could've been registered for multiple webhooks, getting all of them here
	webhooks, err := h.queries.ListWebhooksByEvent(c, req.EventName)

	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "could not list subscriptions")
		return
	}

	// check if it's a valid event
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "data malformed")
		return
	}

	event, err := h.queries.CreateEvent(c, db.CreateEventParams{
		Name: req.EventName,
		Data: dataBytes,
	})

	log.Printf("inserted event ? %v\n", event)

	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "could not save the event")
		return
	}

	var wg sync.WaitGroup

	for _, webhook := range webhooks {
		delivery, err := h.queries.CreateDelivery(c, db.CreateDeliveryParams{
			WebhookID: webhook.ID,
			EventID:   event.ID,
		})

		if err != nil {
			log.Printf("could not create delivery for event %v of webhook %v\n", event.ID, webhook.ID)
			continue
		}

		wg.Go(func() {
			sendWebhook(c.Request.Context(), h, webhook.TargetUrl, req.Data, delivery.ID)
		})
	}

	wg.Wait()

	OK(c, req.EventName)
}
