package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newJSONResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestGenerateContent_DefaultModelAndThoughtFiltering(t *testing.T) {
	var gotPath string
	var gotAPIKey string
	var gotReq generateContentRequest

	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotPath = r.URL.Path
			gotAPIKey = r.Header.Get("x-goog-api-key")
			if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			return newJSONResponse(http.StatusOK, `{
				"candidates": [{
					"content": {
						"role": "model",
						"parts": [
							{"text": "internal thinking...", "thought": true},
							{"text": "Node looks healthy."}
						]
					},
					"finishReason": "STOP"
				}],
				"usageMetadata": {
					"promptTokenCount": 12,
					"candidatesTokenCount": 5,
					"totalTokenCount": 17
				}
			}`), nil
		}),
	}

	client, err := New(
		WithAPIKey("test-key"),
		WithBaseURL("https://example.test"),
		WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := client.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a network engineer."),
		llms.TextParts(llms.ChatMessageTypeHuman, "Diagnose router"),
	}, llms.WithTemperature(0.2))
	if err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}

	if gotPath != "/v1beta/models/"+DefaultModel+":generateContent" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotAPIKey != "test-key" {
		t.Fatalf("unexpected api key header: %s", gotAPIKey)
	}
	if gotReq.SystemInstruction == nil || len(gotReq.SystemInstruction.Parts) != 1 {
		t.Fatalf("expected systemInstruction, got %+v", gotReq.SystemInstruction)
	}
	if gotReq.GenerationConfig == nil || gotReq.GenerationConfig.Temperature == nil || *gotReq.GenerationConfig.Temperature != 0.2 {
		t.Fatalf("unexpected generationConfig: %+v", gotReq.GenerationConfig)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Content != "Node looks healthy." {
		t.Fatalf("unexpected response choices: %+v", resp.Choices)
	}
}

func TestGenerateContent_GemmaSystemFallbackAndModelPrefix(t *testing.T) {
	var gotPath string
	var gotReq generateContentRequest

	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotPath = r.URL.Path
			if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			return newJSONResponse(http.StatusOK, `{
				"candidates": [{
					"content": {"role": "model", "parts": [{"text": "OK"}]},
					"finishReason": "STOP"
				}]
			}`), nil
		}),
	}

	client, err := New(
		WithAPIKey("test-key"),
		WithBaseURL("https://example.test"),
		WithModel("models/gemma-4-26b-a4b-it"),
		WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	resp, err := client.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "System prompt"),
		llms.TextParts(llms.ChatMessageTypeHuman, "User prompt"),
	})
	if err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}

	if gotPath != "/v1beta/models/gemma-4-26b-a4b-it:generateContent" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotReq.SystemInstruction != nil {
		t.Fatalf("expected systemInstruction to be nil for gemma, got %+v", gotReq.SystemInstruction)
	}
	if len(gotReq.Contents) != 1 || len(gotReq.Contents[0].Parts) != 2 {
		t.Fatalf("expected merged system+user parts, got %+v", gotReq.Contents)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Content != "OK" {
		t.Fatalf("unexpected response: %+v", resp.Choices)
	}
}

func TestGenerateContent_APIError(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return newJSONResponse(http.StatusServiceUnavailable, `{"error":{"code":503,"message":"model overloaded","status":"UNAVAILABLE"}}`), nil
		}),
	}

	client, err := New(
		WithAPIKey("test-key"),
		WithBaseURL("https://example.test"),
		WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hello"),
	})
	if err == nil || !strings.Contains(err.Error(), "503") || !strings.Contains(err.Error(), "model overloaded") {
		t.Fatalf("expected formatted 503 error, got: %v", err)
	}
}
