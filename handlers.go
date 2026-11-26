package main

import (
	"encoding/json"
	"log"
	"net/http"
	"webhook-delivery-backend/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
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
		return
	}

	OK(c, webhook)
}

type HandleEventRequest struct {
	EventName string `json:"event_name" binding:"required"`
	Data      any    `json:"data" binding:"required"`
}

func (h *WebhookHandler) HandleEvent(c *gin.Context) {
	var req HandleEventRequest

	// bind req parameters to req
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "STATUS_BAD_REQUEST", "event_name and data are required")
		return
	}

	// read 'data'
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "data malformed")
		return
	}

	ctx := c.Request.Context()

	tx, err := h.pool.Begin(ctx)

	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Something went wrong")
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	webhooks, err := qtx.ListWebhooksByEvent(c, req.EventName)

	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "could not list subscriptions")
		return
	}

	event, err := qtx.CreateEvent(c, db.CreateEventParams{
		Name: req.EventName,
		Data: dataBytes,
	})

	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "could not save the event")
		return
	}

	for _, webhook := range webhooks {
		_, err := qtx.CreateDelivery(c, db.CreateDeliveryParams{
			WebhookID: webhook.ID,
			EventID:   event.ID,
		})

		if err != nil {
			log.Printf("could not create delivery for event %v of webhook %v\n", event.ID, webhook.ID)
			continue
		}
	}

	if err := tx.Commit(ctx); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Something went wrong")
		log.Printf("could not commit the transaction\n")
		return
	}

	Accepted(c, req.EventName)
}
