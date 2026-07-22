package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAnthropicClientCountTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		if request.Header.Get("x-api-key") != "secret" {
			t.Errorf("x-api-key header = %q", request.Header.Get("x-api-key"))
		}
		if request.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version header = %q", request.Header.Get("anthropic-version"))
		}
		var body countTokensRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "test-model" || len(body.Messages) != 1 || body.Messages[0].Content != "hello" {
			t.Errorf("request body = %+v", body)
		}
		writer.Header().Set("content-type", "application/json")
		_, _ = writer.Write([]byte(`{"input_tokens":42}`))
	}))
	defer server.Close()

	client := anthropicClient{
		apiKey:     "secret",
		apiVersion: "2023-06-01",
		endpoint:   server.URL,
		maxRetries: 0,
		httpClient: server.Client(),
	}
	got, err := client.countTokens(context.Background(), "test-model", "hello")
	if err != nil {
		t.Fatalf("countTokens: %v", err)
	}
	if got != 42 {
		t.Errorf("countTokens = %d, want 42", got)
	}
}

func TestAnthropicClientRetriesRateLimit(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			writer.Header().Set("retry-after", "0")
			writer.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = writer.Write([]byte(`{"input_tokens":9}`))
	}))
	defer server.Close()

	client := anthropicClient{
		apiKey:     "secret",
		apiVersion: "2023-06-01",
		endpoint:   server.URL,
		maxRetries: 1,
		httpClient: server.Client(),
	}
	got, err := client.countTokens(context.Background(), "test-model", "hello")
	if err != nil {
		t.Fatalf("countTokens: %v", err)
	}
	if got != 9 || calls.Load() != 2 {
		t.Errorf("got count=%d calls=%d, want count=9 calls=2", got, calls.Load())
	}
}

func TestRetryDelay(t *testing.T) {
	if got := retryDelay("2", 0); got != 2*time.Second {
		t.Errorf("retryDelay seconds = %s, want 2s", got)
	}
	if got := retryDelay("", 0); got != 250*time.Millisecond {
		t.Errorf("retryDelay fallback = %s, want 250ms", got)
	}
}
