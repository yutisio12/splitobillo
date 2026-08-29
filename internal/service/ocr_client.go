package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"splitobillo/internal/config"
)

// OCRService abstracts the OCR engine so it can be swapped without touching
// business logic.
type OCRService interface {
	ExtractReceipt(ctx context.Context, image []byte) (rawText string, err error)
}

type ocrClient struct {
	url    string
	client *http.Client
}

type ocrExtractResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Text string `json:"text"`
	} `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewOCRService(cfg *config.Config) OCRService {
	return &ocrClient{
		url:    strings.TrimRight(cfg.OCRServiceURL, "/"),
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *ocrClient) ExtractReceipt(ctx context.Context, image []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/extract", bytes.NewReader(image))
	if err != nil {
		return "", fmt.Errorf("build ocr request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call ocr service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return "", fmt.Errorf("read ocr response: %w", err)
	}

	var parsed ocrExtractResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode ocr response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !parsed.Success {
		msg := "unknown error"
		if parsed.Error != nil {
			msg = parsed.Error.Message
		}
		return "", fmt.Errorf("ocr failed (status %d): %s", resp.StatusCode, msg)
	}

	return parsed.Data.Text, nil
}
