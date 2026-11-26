package notify

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms"
)

type stubLLM struct {
	fn func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error)
}

func (s *stubLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", nil
}

func (s *stubLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return s.fn(ctx, messages, options...)
}

func TestGenerateSummaryWithRetry_TemporaryErrorRecovers(t *testing.T) {
	attempts := 0
	var slept []time.Duration
	llm := &stubLLM{
		fn: func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("gemini API error 503 (UNAVAILABLE): This model is currently experiencing high demand.")
			}
			return &llms.ContentResponse{
				Choices: []*llms.ContentChoice{{Content: "Summary OK"}},
			}, nil
		},
	}

	resp, err := generateSummaryWithRetry(
		context.Background(),
		llm,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "report")},
		reportLLMRetryDelays,
		func(ctx context.Context, d time.Duration) error {
			slept = append(slept, d)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("generateSummaryWithRetry: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if len(slept) != 2 || slept[0] != 1*time.Minute || slept[1] != 2*time.Minute {
		t.Fatalf("unexpected sleep durations: %v", slept)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Content != "Summary OK" {
		t.Fatalf("unexpected response: %+v", resp.Choices)
	}
}

func TestGenerateSummaryWithRetry_NonTemporaryErrorDoesNotRetry(t *testing.T) {
	attempts := 0
	var slept []time.Duration
	llm := &stubLLM{
		fn: func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
			attempts++
			return nil, errors.New("gemini API error 401 (UNAUTHENTICATED): invalid API key")
		},
	}

	_, err := generateSummaryWithRetry(
		context.Background(),
		llm,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "report")},
		reportLLMRetryDelays,
		func(ctx context.Context, d time.Duration) error {
			slept = append(slept, d)
			return nil
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
	if len(slept) != 0 {
		t.Fatalf("expected no sleeps, got %v", slept)
	}
}
