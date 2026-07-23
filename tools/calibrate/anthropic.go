package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxRetryDelay = 8 * time.Second

type anthropicClient struct {
	apiKey     string
	apiVersion string
	endpoint   string
	maxRetries int
	httpClient *http.Client
}

type countTokensRequest struct {
	Model    string         `json:"model"`
	Messages []inputMessage `json:"messages"`
}

type inputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type countTokensResponse struct {
	InputTokens int64 `json:"input_tokens"`
}

func (c *anthropicClient) countTokens(ctx context.Context, model, content string) (int64, error) {
	body, err := json.Marshal(countTokensRequest{
		Model: model,
		Messages: []inputMessage{{
			Role:    "user",
			Content: content,
		}},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal count_tokens request: %w", err)
	}

	for attempt := 0; ; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
		if err != nil {
			return 0, fmt.Errorf("create count_tokens request: %w", err)
		}
		request.Header.Set("content-type", "application/json")
		request.Header.Set("x-api-key", c.apiKey)
		request.Header.Set("anthropic-version", c.apiVersion)
		request.Header.Set("user-agent", "tloc-calibrate/1")

		response, err := c.httpClient.Do(request)
		if err != nil {
			if attempt < c.maxRetries {
				if err := waitForRetry(ctx, retryDelay("", attempt)); err != nil {
					return 0, err
				}
				continue
			}
			return 0, fmt.Errorf("call count_tokens for model %q: %w", model, err)
		}

		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
		closeErr := response.Body.Close()
		if readErr != nil {
			return 0, fmt.Errorf("read count_tokens response: %w", readErr)
		}
		if closeErr != nil {
			return 0, fmt.Errorf("close count_tokens response: %w", closeErr)
		}

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			var result countTokensResponse
			if err := json.Unmarshal(responseBody, &result); err != nil {
				return 0, fmt.Errorf("decode count_tokens response: %w", err)
			}
			if result.InputTokens < 0 {
				return 0, fmt.Errorf("count_tokens returned negative input_tokens %d", result.InputTokens)
			}
			return result.InputTokens, nil
		}

		if isRetryableStatus(response.StatusCode) && attempt < c.maxRetries {
			if err := waitForRetry(ctx, retryDelay(response.Header.Get("retry-after"), attempt)); err != nil {
				return 0, err
			}
			continue
		}

		detail := strings.TrimSpace(string(responseBody))
		if len(detail) > 2000 {
			detail = detail[:2000] + "..."
		}
		return 0, fmt.Errorf("count_tokens for model %q returned %s: %s", model, response.Status, detail)
	}
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusRequestTimeout || status >= 500
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	return retryDelayAt(retryAfter, attempt, time.Now())
}

func retryDelayAt(retryAfter string, attempt int, now time.Time) time.Duration {
	retryAfter = strings.TrimSpace(retryAfter)
	if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil && seconds >= 0 {
		if seconds >= int64(maxRetryDelay/time.Second) {
			return maxRetryDelay
		}
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(retryAfter); err == nil {
		if delay := when.Sub(now); delay > 0 {
			if delay > maxRetryDelay {
				return maxRetryDelay
			}
			return delay
		}
	}
	delay := 250 * time.Millisecond * time.Duration(1<<min(attempt, 5))
	if delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
