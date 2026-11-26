package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

const (
	// DefaultBaseURL is the default Gemini API endpoint base URL.
	DefaultBaseURL = "https://generativelanguage.googleapis.com"
	// DefaultModel is used when no model is explicitly configured.
	DefaultModel = "gemini-3.5-flash-lite"
	// maxErrorBodyBytes limits raw response body snippets in error messages.
	maxErrorBodyBytes     = 512
	defaultMaxRetries     = 3
	defaultRetryBaseDelay = 2 * time.Second
	maxRetryDelay         = 30 * time.Second
)

// Client implements langchaingo's llms.Model using the Gemini REST API.
type Client struct {
	apiKey         string
	baseURL        string
	model          string
	httpClient     *http.Client
	maxRetries     int
	retryBaseDelay time.Duration
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the Gemini API key.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = strings.TrimSpace(apiKey)
	}
}

// WithBaseURL overrides the Gemini API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/"); trimmed != "" {
			c.baseURL = trimmed
		}
	}
}

// WithModel sets the default model name.
func WithModel(model string) Option {
	return func(c *Client) {
		if trimmed := normalizeModelName(model); trimmed != "" {
			c.model = trimmed
		}
	}
}

// WithHTTPClient sets a custom HTTP client (primarily for testing).
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// New creates a new Gemini REST API client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:        DefaultBaseURL,
		model:          DefaultModel,
		httpClient:     http.DefaultClient,
		maxRetries:     defaultMaxRetries,
		retryBaseDelay: defaultRetryBaseDelay,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.apiKey == "" {
		return nil, errors.New("gemini API key is empty")
	}
	return c, nil
}

// Call implements llms.Model.
func (c *Client) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, c, prompt, options...)
}

// GenerateContent implements llms.Model.
func (c *Client) GenerateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	opts := llms.CallOptions{
		Model: c.model,
	}
	for _, opt := range options {
		opt(&opts)
	}

	modelName := normalizeModelName(opts.Model)
	if modelName == "" {
		modelName = c.model
	}
	if modelName == "" {
		modelName = DefaultModel
	}

	reqBody, err := buildRequest(modelName, messages, &opts)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.baseURL, modelName)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		resp, statusCode, retryAfter, err := c.doGenerateContentRequest(ctx, endpoint, payload)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if attempt == c.maxRetries || !isRetryableStatus(statusCode) {
			break
		}
		delay := c.computeRetryDelay(attempt, retryAfter)
		if sleepErr := sleepWithContext(ctx, delay); sleepErr != nil {
			return nil, lastErr
		}
	}
	return nil, lastErr
}

func (c *Client) doGenerateContentRequest(ctx context.Context, endpoint string, payload []byte) (*llms.ContentResponse, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", fmt.Errorf("failed to read gemini response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, resp.Header.Get("Retry-After"), formatAPIError(resp.StatusCode, body)
	}

	var apiResp generateContentResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, resp.StatusCode, "", fmt.Errorf("failed to decode gemini response: %w (body: %s)", err, truncateBody(body))
	}
	if apiResp.Error != nil {
		return nil, apiResp.Error.Code, "", fmt.Errorf("gemini API error (%d %s): %s", apiResp.Error.Code, apiResp.Error.Status, apiResp.Error.Message)
	}
	if len(apiResp.Candidates) == 0 {
		return nil, resp.StatusCode, "", errors.New("no content in gemini response")
	}

	return convertResponse(&apiResp), resp.StatusCode, "", nil
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (c *Client) computeRetryDelay(attempt int, retryAfter string) time.Duration {
	if c.retryBaseDelay <= 0 {
		return 0
	}
	if sec, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && sec > 0 {
		d := time.Duration(sec) * time.Second
		if d > maxRetryDelay {
			return maxRetryDelay
		}
		return d
	}
	delay := c.retryBaseDelay << attempt
	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}
	// Add +/- 20% jitter to avoid synchronized retries.
	jitterRange := delay / 5
	if jitterRange > 0 {
		delay = delay - jitterRange + time.Duration(rand.Int64N(int64(jitterRange*2)+1))
	}
	return delay
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func normalizeModelName(model string) string {
	trimmed := strings.TrimSpace(model)
	return strings.TrimPrefix(trimmed, "models/")
}

func buildRequest(modelName string, messages []llms.MessageContent, opts *llms.CallOptions) (*generateContentRequest, error) {
	var systemTexts []string
	contents := make([]content, 0, len(messages))

	for _, msg := range messages {
		if msg.Role == llms.ChatMessageTypeSystem {
			for _, p := range msg.Parts {
				if tp, ok := p.(llms.TextContent); ok && tp.Text != "" {
					systemTexts = append(systemTexts, tp.Text)
				}
			}
			continue
		}

		parts, err := convertParts(msg.Parts)
		if err != nil {
			return nil, err
		}
		if len(parts) == 0 {
			continue
		}

		role := "user"
		if msg.Role == llms.ChatMessageTypeAI {
			role = "model"
		}
		contents = append(contents, content{
			Role:  role,
			Parts: parts,
		})
	}

	req := &generateContentRequest{
		GenerationConfig: buildGenerationConfig(opts),
	}

	// Gemma models reject systemInstruction; prepend system text to the first user message.
	if len(systemTexts) > 0 {
		joinedSystem := strings.Join(systemTexts, "\n")
		if isGemmaModel(modelName) || len(contents) == 0 {
			if len(contents) == 0 {
				contents = append(contents, content{
					Role:  "user",
					Parts: []part{{Text: joinedSystem}},
				})
			} else if contents[0].Role == "user" {
				contents[0].Parts = append([]part{{Text: joinedSystem + "\n\n"}}, contents[0].Parts...)
			} else {
				contents = append([]content{{
					Role:  "user",
					Parts: []part{{Text: joinedSystem}},
				}}, contents...)
			}
		} else {
			req.SystemInstruction = &content{
				Parts: []part{{Text: joinedSystem}},
			}
		}
	}

	if len(contents) == 0 {
		return nil, errors.New("no messages to send to gemini")
	}
	req.Contents = contents
	return req, nil
}

func isGemmaModel(modelName string) bool {
	return strings.HasPrefix(strings.ToLower(modelName), "gemma")
}

func convertParts(parts []llms.ContentPart) ([]part, error) {
	out := make([]part, 0, len(parts))
	for _, p := range parts {
		switch v := p.(type) {
		case llms.TextContent:
			out = append(out, part{Text: v.Text})
		case llms.BinaryContent:
			out = append(out, part{
				InlineData: &inlineData{
					MIMEType: v.MIMEType,
					Data:     base64.StdEncoding.EncodeToString(v.Data),
				},
			})
		default:
			return nil, fmt.Errorf("unsupported content part type %T", p)
		}
	}
	return out, nil
}

func buildGenerationConfig(opts *llms.CallOptions) *generationConfig {
	if opts == nil {
		return nil
	}
	cfg := &generationConfig{}
	hasField := false

	if opts.Temperature > 0 {
		t := opts.Temperature
		cfg.Temperature = &t
		hasField = true
	}
	if opts.TopP > 0 {
		p := opts.TopP
		cfg.TopP = &p
		hasField = true
	}
	if opts.TopK > 0 {
		k := opts.TopK
		cfg.TopK = &k
		hasField = true
	}
	if opts.MaxTokens > 0 {
		m := opts.MaxTokens
		cfg.MaxOutputTokens = &m
		hasField = true
	}
	if opts.CandidateCount > 0 {
		c := opts.CandidateCount
		cfg.CandidateCount = &c
		hasField = true
	}
	if len(opts.StopWords) > 0 {
		cfg.StopSequences = opts.StopWords
		hasField = true
	}
	if opts.ResponseMIMEType != "" {
		cfg.ResponseMIMEType = opts.ResponseMIMEType
		hasField = true
	} else if opts.JSONMode {
		cfg.ResponseMIMEType = "application/json"
		hasField = true
	}

	if !hasField {
		return nil
	}
	return cfg
}

func convertResponse(apiResp *generateContentResponse) *llms.ContentResponse {
	choices := make([]*llms.ContentChoice, 0, len(apiResp.Candidates))
	for _, cand := range apiResp.Candidates {
		var answer strings.Builder
		var fallbackThought strings.Builder
		if cand.Content != nil {
			for _, p := range cand.Content.Parts {
				if p.Text == "" {
					continue
				}
				if p.Thought {
					fallbackThought.WriteString(p.Text)
					continue
				}
				answer.WriteString(p.Text)
			}
		}
		text := answer.String()
		if text == "" {
			text = fallbackThought.String()
		}

		meta := make(map[string]any)
		if apiResp.UsageMetadata != nil {
			meta["input_tokens"] = apiResp.UsageMetadata.PromptTokenCount
			meta["output_tokens"] = apiResp.UsageMetadata.CandidatesTokenCount
			meta["total_tokens"] = apiResp.UsageMetadata.TotalTokenCount
			meta["PromptTokens"] = apiResp.UsageMetadata.PromptTokenCount
			meta["CompletionTokens"] = apiResp.UsageMetadata.CandidatesTokenCount
			meta["TotalTokens"] = apiResp.UsageMetadata.TotalTokenCount
		}

		choices = append(choices, &llms.ContentChoice{
			Content:        text,
			StopReason:     cand.FinishReason,
			GenerationInfo: meta,
		})
	}
	return &llms.ContentResponse{Choices: choices}
}

func formatAPIError(statusCode int, body []byte) error {
	var errResp struct {
		Error *apiError `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		return fmt.Errorf("gemini API error %d (%s): %s", statusCode, errResp.Error.Status, errResp.Error.Message)
	}
	return fmt.Errorf("gemini API HTTP %d: %s", statusCode, truncateBody(body))
}

func truncateBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > maxErrorBodyBytes {
		return s[:maxErrorBodyBytes] + "..."
	}
	return s
}

type generateContentRequest struct {
	Contents          []content         `json:"contents"`
	SystemInstruction *content          `json:"systemInstruction,omitempty"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
	Thought    bool        `json:"thought,omitempty"`
}

type inlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

type generationConfig struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
	TopK             *int     `json:"topK,omitempty"`
	MaxOutputTokens  *int     `json:"maxOutputTokens,omitempty"`
	CandidateCount   *int     `json:"candidateCount,omitempty"`
	StopSequences    []string `json:"stopSequences,omitempty"`
	ResponseMIMEType string   `json:"responseMimeType,omitempty"`
}

type generateContentResponse struct {
	Candidates    []candidate    `json:"candidates"`
	UsageMetadata *usageMetadata `json:"usageMetadata,omitempty"`
	Error         *apiError      `json:"error,omitempty"`
}

type candidate struct {
	Content      *content `json:"content,omitempty"`
	FinishReason string   `json:"finishReason,omitempty"`
}

type usageMetadata struct {
	PromptTokenCount     int32 `json:"promptTokenCount"`
	CandidatesTokenCount int32 `json:"candidatesTokenCount"`
	TotalTokenCount      int32 `json:"totalTokenCount"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
