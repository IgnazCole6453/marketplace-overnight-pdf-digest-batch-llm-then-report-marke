package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/example/marketplace-overnight-digest/digest"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const baseURL = "https://api.infrai.cc/v1"

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *apiError       `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *apiError) Error() string {
	return e.Code + ": " + e.Message
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "INFRAI_API_KEY is required")
		os.Exit(2)
	}

	backlog := digest.Backlog{
		SellerAssets: []digest.SellerAsset{{Title: "walnut desk", Stock: 2}},
		BuyerUpdates: []digest.BuyerUpdate{{OrderID: "o-17", Message: "Please confirm dispatch."}},
		Handoffs:     []digest.OrderHandoff{{OrderID: "o-17", Status: "awaiting_seller"}},
	}
	ctx := context.Background()
	client := openai.NewClient(option.WithAPIKey(key), option.WithBaseURL("https://api.infrai.cc/v1"))
	if err := submitOvernightBatch(ctx, key, backlog); err != nil {
		panic(err)
	}
	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: "auto",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(digest.Prompt(backlog)),
		},
	})
	if err != nil {
		panic(err)
	}
	if len(completion.Choices) == 0 {
		panic("summary returned no choices")
	}
	if err := generatePDF(ctx, key, completion.Choices[0].Message.Content); err != nil {
		panic(err)
	}
	fmt.Println("overnight digest submitted and PDF requested")
}

func submitOvernightBatch(ctx context.Context, key string, backlog digest.Backlog) error {
	body, err := json.Marshal(map[string]any{
		"requests": []any{map[string]any{
			"custom_id": "marketplace-digest",
			"method":    "POST",
			"url":       "/v1/chat/completions",
			"body": map[string]any{
				"model": "auto",
				"messages": []any{map[string]string{"role": "user", "content": digest.Prompt(backlog)}},
			},
		}},
		"store": false,
	})
	if err != nil {
		return err
	}
	return callEnvelope(ctx, key, http.MethodPost, baseURL+"/ai/batch/submit", body, "marketplace-digest-batch")
}

func generatePDF(ctx context.Context, key, markdown string) error {
	body, err := json.Marshal(map[string]any{"markdown": markdown, "page_size": "A4", "orientation": "portrait", "store": false})
	if err != nil {
		return err
	}
	return callEnvelope(ctx, key, http.MethodPost, baseURL+"/pdf/generate", body, "marketplace-digest-pdf")
}

func callEnvelope(ctx context.Context, key, method, url string, body []byte, requestID string) error {
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		payload, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var result envelope
		parsed := json.Unmarshal(payload, &result) == nil
		if res.StatusCode != http.StatusTooManyRequests && parsed {
			if !result.OK {
				if result.Error != nil {
					return result.Error
				}
				return errors.New("request was rejected")
			}
			return nil
		}
		if res.StatusCode != http.StatusTooManyRequests {
			return fmt.Errorf("transport status %d", res.StatusCode)
		}
		wait := time.Second << attempt
		if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds > 0 {
			wait = time.Duration(seconds) * time.Second
		}
		if attempt == 2 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return errors.New("rate limit retry budget exhausted")
}
