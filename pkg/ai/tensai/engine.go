package tensai

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mattn/tensai/gpu"
	"github.com/mattn/tensai/tokenizer"
)

// Engine runs autoregressive text generation using tensai with GPU/SIMD/CPU support
type Engine struct {
	model *qwen
	gpu   *gpuQwen
	tok   *tokenizer.Tokenizer
	imEnd int
	eot   int
	rng   *rand.Rand
	mu    sync.Mutex
}

// LoadEngine loads the model from GGUF and attempts to initialize GPU acceleration
func LoadEngine(modelPath string) (*Engine, error) {
	return LoadEngineWithOptions(modelPath, false)
}

// LoadEngineWithOptions loads the model with optional GPU disabling
func LoadEngineWithOptions(modelPath string, noGPU bool) (*Engine, error) {
	var useGPU bool
	var dev *gpu.Device
	if !noGPU {
		d, devErr := GetGPUDevice()
		if devErr == nil && d != nil {
			dev = d
			useGPU = true
		}
	}

	// Load 8-bit quantized weights for memory-efficient and fast execution
	model, tok, err := loadQwenGGUF(modelPath, 8, useGPU)
	if err != nil {
		return nil, err
	}

	imEnd := model.cfg.EOS
	if id, ok := tok.ID("<|im_end|>"); ok {
		imEnd = id
	}
	eot := model.cfg.EOS
	if id, ok := tok.ID("<|endoftext|>"); ok {
		eot = id
	}

	engine := &Engine{
		model: model,
		tok:   tok,
		imEnd: imEnd,
		eot:   eot,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Attempt GPU acceleration (WebGPU / Metal / Vulkan / D3D12)
	if useGPU {
		nCtx := model.cfg.MaxPos
		if nCtx <= 0 {
			nCtx = 8192
		}
		kvDim := model.cfg.KVHeads * model.headSz
		if kvDim > 0 && model.cfg.Layers > 0 {
			// Budget up to 2GB for KV caches
			maxCtx := (2 << 30) / (2 * model.cfg.Layers * kvDim * 4)
			if lim := dev.StorageLimit(); lim > 0 {
				if perBuf := int(lim / uint64(kvDim*4)); perBuf > 0 && perBuf < maxCtx {
					maxCtx = perBuf
				}
			}
			if maxCtx > 0 && maxCtx < nCtx {
				nCtx = maxCtx
			}
		}
		if gq, err := newGPUQwen(model, dev, nCtx); err == nil {
			engine.gpu = gq
		}
	}

	return engine, nil
}

func sampleToken(logits []float32, temp float64, topP float64, rng *rand.Rand) int {
	if temp <= 0 {
		bestID := 0
		bestVal := logits[0]
		for i := 1; i < len(logits); i++ {
			if logits[i] > bestVal {
				bestVal = logits[i]
				bestID = i
			}
		}
		return bestID
	}

	var maxLogit float32 = -1e9
	for _, l := range logits {
		if l > maxLogit {
			maxLogit = l
		}
	}

	probs := make([]float64, len(logits))
	var sum float64
	invT := 1.0 / temp
	for i, l := range logits {
		p := math.Exp(float64(l-maxLogit) * invT)
		probs[i] = p
		sum += p
	}

	r := rng.Float64() * sum
	var acc float64
	for i, p := range probs {
		acc += p
		if acc >= r {
			return i
		}
	}
	return len(logits) - 1
}

// GenerateText streams generated tokens
func (e *Engine) GenerateText(ctx context.Context, prompt string, maxTokens int, streamingFunc func(ctx context.Context, chunk []byte) error) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.model.reset()
	if e.gpu != nil {
		e.gpu.reset()
	}

	limit := maxTokens
	if limit <= 0 {
		limit = 512
	}

	formattedPrompt := fmt.Sprintf("<|im_start|>system\nYou are a helpful network management and IT assistant. Follow the user's instructions and language request carefully.<|im_end|>\n<|im_start|>user\n%s<|im_end|>\n<|im_start|>assistant\n", prompt)
	ids := e.tok.Encode(formattedPrompt)
	if len(ids) == 0 {
		ids = e.tok.Encode(prompt)
	}

	// Safety check: ensure prompt tokens fit within model/GPU context capacity
	maxCtxCap := e.model.cfg.MaxPos
	if e.gpu != nil && e.gpu.nCtx < maxCtxCap {
		maxCtxCap = e.gpu.nCtx
	}
	if maxCtxCap <= 0 {
		maxCtxCap = 4096
	}
	availForPrompt := maxCtxCap - limit - 16
	if availForPrompt < 128 {
		availForPrompt = 128
	}
	if len(ids) > availForPrompt {
		// Truncate to fit within context, keeping the tail end of the prompt
		ids = ids[len(ids)-availForPrompt:]
	}

	// Prefill / feed prompt
	var logits []float32
	if e.gpu != nil {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		logits = e.gpu.prefill(ids, 0)
	} else {
		for pos, id := range ids {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}
			logits = e.model.step(id, pos)
		}
	}

	// Decode loop
	var result strings.Builder
	pos := len(ids)

	for i := 0; i < limit && pos < e.model.cfg.MaxPos-1; i++ {
		select {
		case <-ctx.Done():
			return result.String(), ctx.Err()
		default:
		}

		next := sampleToken(logits, 0.0, 0.9, e.rng)
		if next == e.imEnd || next == e.eot {
			break
		}

		piece := e.tok.Decode([]int{next})
		result.WriteString(piece)

		if streamingFunc != nil {
			if err := streamingFunc(ctx, []byte(piece)); err != nil {
				return result.String(), err
			}
		}

		if e.gpu != nil {
			logits = e.gpu.step(next, pos)
		} else {
			logits = e.model.step(next, pos)
		}
		pos++
	}

	return result.String(), nil
}
