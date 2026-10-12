package tensai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// TensaiLLM implements langchaingo's llms.Model for embedded inference using tensai
type TensaiLLM struct {
	modelPath string
	engine    *Engine
	accType   AccelerationType
	accDetail string
}

// New creates a new TensaiLLM instance with automatic GPU -> SIMD -> CPU fallback
func New(modelPath string) (*TensaiLLM, error) {
	return NewWithOptions(modelPath, false)
}

// NewWithOptions creates a new TensaiLLM instance with optional GPU bypass
func NewWithOptions(modelPath string, noGPU bool) (*TensaiLLM, error) {
	if modelPath == "" {
		return nil, errors.New("model path cannot be empty")
	}

	accType, accDetail := DetectAccelerationWithOptions(noGPU)

	// Load neural engine with GPU if available, or fallback to CPU
	engine, err := LoadEngineWithOptions(modelPath, noGPU)
	if err != nil {
		log.Printf("TensaiLLM: LoadEngineWithOptions failed (model=%s): %v", modelPath, err)
		return nil, fmt.Errorf("failed to load local LLM model %q: %w", modelPath, err)
	}

	return &TensaiLLM{
		modelPath: modelPath,
		engine:    engine,
		accType:   accType,
		accDetail: accDetail,
	}, nil
}

// Acceleration returns the current active acceleration type and hardware details
func (m *TensaiLLM) Acceleration() (AccelerationType, string) {
	if m.engine != nil && m.engine.gpu != nil {
		return AccelGPU, m.accDetail
	}
	return m.accType, m.accDetail
}

// Call generates text from a prompt
func (m *TensaiLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	resp, err := m.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices returned")
	}
	return resp.Choices[0].Content, nil
}

// GenerateContent generates responses for structured chat messages
func (m *TensaiLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if m.engine == nil {
		err := errors.New("tensai LLM engine is not initialized")
		log.Printf("TensaiLLM: GenerateContent error: %v", err)
		return nil, err
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Extract prompt text from messages
	var sb strings.Builder
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				sb.WriteString(textPart.Text)
				sb.WriteString("\n")
			}
		}
	}
	prompt := strings.TrimSpace(sb.String())

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 512
	}
	responseContent, err := m.engine.GenerateText(ctx, prompt, maxTokens, opts.StreamingFunc)
	if err != nil {
		log.Printf("TensaiLLM: GenerateText error: %v", err)
		return nil, fmt.Errorf("tensai inference error: %w", err)
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: responseContent,
			},
		},
	}, nil
}
