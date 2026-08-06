package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestOpenAICompatibleGeneratesStructuredOutput(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		var request chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.ResponseFormat.Type != "json_object" {
			t.Fatalf("response format = %q", request.ResponseFormat.Type)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"title\":\"结构化结果\"}"}}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewOpenAICompatible(OpenAICompatibleConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-key",
		Model:   "test-model",
		Timeout: time.Second,
	})
	var output struct {
		Title string `json:"title"`
	}
	metadata, err := provider.Generate(t.Context(), Prompt{Version: "prompt-v1", Instruction: "JSON"}, map[string]string{"text": "hello"}, &output)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if output.Title != "结构化结果" {
		t.Fatalf("output title = %q", output.Title)
	}
	if metadata.Model != "test-model" || metadata.PromptVersion != "prompt-v1" {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestOpenAICompatibleRetriesTransientFailure(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"ok\":true}"}}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewOpenAICompatible(OpenAICompatibleConfig{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		Model:      "test-model",
		Timeout:    time.Second,
		MaxRetries: 1,
	})
	provider.sleep = func(context.Context, time.Duration) error { return nil }
	var output struct {
		OK bool `json:"ok"`
	}
	if _, err := provider.Generate(t.Context(), Prompt{}, struct{}{}, &output); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !output.OK || attempts.Load() != 2 {
		t.Fatalf("output = %#v, attempts = %d", output, attempts.Load())
	}
}

func TestOpenAICompatibleRejectsMalformedStructuredOutput(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"not-json"}}]}`))
	}))
	t.Cleanup(server.Close)
	provider := NewOpenAICompatible(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
		Timeout: time.Second,
	})

	var output map[string]any
	_, err := provider.Generate(t.Context(), Prompt{}, struct{}{}, &output)
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("Generate() error = %v, want ErrInvalidResponse", err)
	}
}

func TestOpenAICompatibleHonorsTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(80 * time.Millisecond)
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	t.Cleanup(server.Close)
	provider := NewOpenAICompatible(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
		Timeout: 10 * time.Millisecond,
	})

	var output map[string]any
	_, err := provider.Generate(t.Context(), Prompt{}, struct{}{}, &output)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Generate() error = %v, want deadline exceeded", err)
	}
}

func TestPromptRegistryIncludesGuardedTasks(t *testing.T) {
	t.Parallel()
	for _, task := range []string{
		TaskIntentClarification,
		TaskMatchExplanation,
		TaskConversationStarters,
		TaskGroupSummary,
	} {
		prompt, ok := PromptFor(task)
		if !ok || prompt.Version == "" || prompt.Instruction == "" {
			t.Fatalf("prompt %q = %#v, %v", task, prompt, ok)
		}
	}
}
