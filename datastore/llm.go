package datastore

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/twsnmp/twsnmpfk/pkg/ai/tensai"
	"github.com/twsnmp/twsnmpfk/pkg/model"
)

func GetLLM(ctx context.Context) (llms.Model, error) {
	switch MapConf.LLMProvider {
	case "tensai", "embedded", "local":
		modelPath, err := model.FindModel("", MapConf.LLMModel)
		if err != nil {
			log.Printf("GetLLM tensai: local model %q not found: %v", MapConf.LLMModel, err)
			return nil, fmt.Errorf("local model not found (%s): %w", MapConf.LLMModel, err)
		}
		llm, err := tensai.NewWithOptions(modelPath, false)
		if err != nil {
			log.Printf("GetLLM tensai: failed to load model %s: %v", modelPath, err)
			return nil, err
		}
		return llm, nil
	case "ollama":
		baseURL := "http://localhost:11434"
		if MapConf.LLMBaseURL != "" {
			baseURL = MapConf.LLMBaseURL
		}
		return ollama.New(
			ollama.WithModel(MapConf.LLMModel),
			ollama.WithServerURL(baseURL),
		)
	case "gemini", "googleai":
		opts := []googleai.Option{
			googleai.WithAPIKey(MapConf.LLMAPIKey),
		}
		if MapConf.LLMModel != "" {
			opts = append(opts, googleai.WithDefaultModel(MapConf.LLMModel))
		}
		return googleai.New(ctx, opts...)
	case "openai":
		opts := []openai.Option{}
		if MapConf.LLMModel != "" {
			opts = append(opts, openai.WithModel(MapConf.LLMModel))
		}
		if MapConf.LLMAPIKey != "" {
			opts = append(opts, openai.WithToken(MapConf.LLMAPIKey))
		}
		if MapConf.LLMBaseURL != "" {
			opts = append(opts, openai.WithBaseURL(MapConf.LLMBaseURL))
		}
		return openai.New(opts...)
	case "anthropic", "claude":
		opts := []anthropic.Option{}
		if MapConf.LLMModel != "" {
			opts = append(opts, anthropic.WithModel(MapConf.LLMModel))
		}
		if MapConf.LLMAPIKey != "" {
			opts = append(opts, anthropic.WithToken(MapConf.LLMAPIKey))
		}
		if MapConf.LLMBaseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(MapConf.LLMBaseURL))
		}
		return anthropic.New(opts...)
	}
	return nil, fmt.Errorf("llm provider not found")
}
