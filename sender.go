package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func sendWebhook(ctx context.Context, targetURL string, payload any, deliveryId int64) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		targetURL,
		bytes.NewReader(body),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %v\n", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("target returned HTTP %d\n", res.StatusCode)
	}

	log.Printf("sendWebhook success, updating delivery %v\n", deliveryId)

	return nil
}
