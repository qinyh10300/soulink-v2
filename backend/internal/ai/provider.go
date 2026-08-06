package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxResponseBytes = 1 << 20

var (
	ErrUnavailable     = errors.New("ai provider unavailable")
	ErrInvalidResponse = errors.New("ai provider returned invalid structured output")
	ErrRequestRejected = errors.New("ai provider rejected request")
)

type Metadata struct {
	Source        string
	Model         string
	PromptVersion string
}

type Generator interface {
	Generate(ctx context.Context, prompt Prompt, input any, output any) (Metadata, error)
}

type OpenAICompatibleConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
	HTTPClient *http.Client
}

type OpenAICompatible struct {
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
	maxRetries int
	client     *http.Client
	sleep      func(context.Context, time.Duration) error
}

func NewOpenAICompatible(config OpenAICompatibleConfig) *OpenAICompatible {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	maxRetries := config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries > 3 {
		maxRetries = 3
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	return &OpenAICompatible{
		baseURL:    strings.TrimRight(strings.TrimSpace(config.BaseURL), "/"),
		apiKey:     strings.TrimSpace(config.APIKey),
		model:      strings.TrimSpace(config.Model),
		timeout:    timeout,
		maxRetries: maxRetries,
		client:     client,
		sleep:      sleepWithContext,
	}
}

func (provider *OpenAICompatible) Enabled() bool {
	return provider != nil && provider.baseURL != "" && provider.apiKey != "" && provider.model != ""
}

func (provider *OpenAICompatible) Generate(
	ctx context.Context,
	prompt Prompt,
	input any,
	output any,
) (Metadata, error) {
	metadata := Metadata{Source: "model", Model: provider.model, PromptVersion: prompt.Version}
	if !provider.Enabled() {
		return metadata, ErrUnavailable
	}
	if output == nil {
		return metadata, fmt.Errorf("%w: nil output", ErrInvalidResponse)
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return metadata, fmt.Errorf("marshal model input: %w", err)
	}
	body, err := json.Marshal(chatCompletionRequest{
		Model: provider.model,
		Messages: []chatMessage{
			{Role: "system", Content: prompt.Instruction},
			{Role: "user", Content: string(inputJSON)},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
		Temperature:    0.2,
	})
	if err != nil {
		return metadata, fmt.Errorf("marshal model request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= provider.maxRetries; attempt++ {
		if attempt > 0 {
			if err := provider.sleep(ctx, time.Duration(1<<(attempt-1))*200*time.Millisecond); err != nil {
				return metadata, err
			}
		}
		content, retry, err := provider.request(ctx, body)
		if err == nil {
			decoder := json.NewDecoder(strings.NewReader(content))
			decoder.DisallowUnknownFields()
			if decodeErr := decoder.Decode(output); decodeErr != nil {
				return metadata, fmt.Errorf("%w: %v", ErrInvalidResponse, decodeErr)
			}
			if decoder.Decode(&struct{}{}) != io.EOF {
				return metadata, fmt.Errorf("%w: trailing JSON data", ErrInvalidResponse)
			}
			return metadata, nil
		}
		lastErr = err
		if !retry {
			break
		}
	}
	return metadata, lastErr
}

func (provider *OpenAICompatible) request(ctx context.Context, body []byte) (string, bool, error) {
	requestCtx, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodPost,
		provider.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", false, fmt.Errorf("create model request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+provider.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := provider.client.Do(req)
	if err != nil {
		if requestCtx.Err() != nil {
			return "", true, fmt.Errorf("model request timeout: %w", requestCtx.Err())
		}
		return "", true, fmt.Errorf("model request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if readErr != nil {
		return "", true, fmt.Errorf("read model response: %w", readErr)
	}
	if len(raw) > maxResponseBytes {
		return "", false, fmt.Errorf("%w: response too large", ErrInvalidResponse)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retry := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return "", retry, fmt.Errorf("%w: status %d", ErrRequestRejected, resp.StatusCode)
	}

	var response chatCompletionResponse
	if err := json.Unmarshal(raw, &response); err != nil || len(response.Choices) == 0 {
		return "", false, fmt.Errorf("%w: malformed completion envelope", ErrInvalidResponse)
	}
	content := strings.TrimSpace(response.Choices[0].Message.Content)
	if content == "" {
		return "", false, fmt.Errorf("%w: empty completion", ErrInvalidResponse)
	}
	return content, false, nil
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
	Temperature    float64        `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
