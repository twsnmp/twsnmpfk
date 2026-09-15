package tensai

import (
	"math"

	tensai "github.com/mattn/tensai"
	"github.com/mattn/tensai/quant"
)

type config struct {
	HiddenSize   int     `json:"hidden_size"`
	HeadDim      int     `json:"head_dim"`
	NoRopeLayers []int   `json:"no_rope_layers"`
	SlidingWin   int     `json:"sliding_window"`
	Intermediate int     `json:"intermediate_size"`
	Layers       int     `json:"num_hidden_layers"`
	Heads        int     `json:"num_attention_heads"`
	KVHeads      int     `json:"num_key_value_heads"`
	RMSEps       float64 `json:"rms_norm_eps"`
	RopeTheta    float64 `json:"rope_theta"`
	MaxPos       int     `json:"max_position_embeddings"`
	Vocab        int     `json:"vocab_size"`
	TieEmbedding bool    `json:"tie_word_embeddings"`
	EOS          int     `json:"eos_token_id"`
	ModelType    string  `json:"model_type"`
	ChatStyle    string  `json:"-"`
	ChatTemplate string  `json:"-"`
	NExpert      int     `json:"-"`
	NExpertUsed  int     `json:"-"`
	MoeFF        int     `json:"-"`
	SharedFF     int     `json:"-"`
	YarnFactor   float64 `json:"-"`
	YarnOrigCtx  int     `json:"-"`
	YarnBetaFast float64 `json:"-"`
	YarnBetaSlow float64 `json:"-"`
}

type qmat struct {
	cols int
	f    func(x, out []float32) error
	mm   func(x, out *tensai.Matrix) error
	q8   *quant.QMatrix
	q4   *quant.Q4Matrix
	q8g  *quant.Q8GMatrix
	mx   *quant.MXFP4Matrix
}

func qmatQ8(q *quant.QMatrix) *qmat {
	return &qmat{cols: q.Cols, f: q.MatVec, mm: q.MatMul, q8: q}
}

func qmatQ4(q *quant.Q4Matrix) *qmat {
	return &qmat{cols: q.Cols, f: q.MatVec, mm: q.MatMul, q4: q}
}

func qmatQ8G(q *quant.Q8GMatrix) *qmat {
	return &qmat{cols: q.Cols, f: q.MatVec, mm: q.MatMul, q8g: q}
}

func quantizeMat(m *tensai.Matrix, bits int) *qmat {
	switch bits {
	case 8:
		return qmatQ8(quant.Quantize(m))
	case 4:
		q, err := quant.Quantize4(m)
		if err != nil {
			panic(err)
		}
		return qmatQ4(q)
	}
	panic("unsupported quantization width")
}

type qblock struct {
	ln1, ln2     []float32
	qNorm, kNorm []float32
	postAttn     []float32
	postFFN      []float32
	noPE         bool
	window       int
	ropeTheta    float64
	ropeFreq     []float64
	geglu        bool
	wQKV, wo     *tensai.Matrix
	bQKV         []float32
	wGU, wDown   *tensai.Matrix
	qQKV, qo     *qmat
	qGU, qDown   *qmat
	router       *tensai.Matrix
	routerBias   []float32
	topK         int
	normTopK     bool
	softmaxK     bool
	oaiGLU       bool
	sinks        []float32
	bo           []float32
	sharedGU     *qmat
	sharedDown   *qmat
	sharedGate   []float32
	kc, vc       [][]float32
}

type qwen struct {
	cfg    config
	headSz int
	embed  *tensai.Tensor
	lmT    *tensai.Matrix
	qLmT   *qmat
	normW  []float32
	blocks []qblock
}

func (m *qwen) initRopeFreqs() {
	half := m.headSz / 2
	for i := range m.blocks {
		b := &m.blocks[i]
		theta := b.ropeTheta
		if theta == 0 {
			theta = m.cfg.RopeTheta
		}
		if theta == 0 {
			theta = 10000
		}
		b.ropeFreq = make([]float64, half)
		for j := 0; j < half; j++ {
			b.ropeFreq[j] = 1.0 / math.Pow(theta, float64(2*j)/float64(m.headSz))
		}
	}
}

func (m *qwen) reset() {
	for i := range m.blocks {
		m.blocks[i].kc = nil
		m.blocks[i].vc = nil
	}
}

func rmsnormInto(out, in, w []float32, eps float64) {
	var sum float64
	for _, v := range in {
		sum += float64(v * v)
	}
	mean := sum / float64(len(in))
	scale := float32(1.0 / math.Sqrt(mean+eps))
	for i, v := range in {
		factor := float32(1.0)
		if i < len(w) {
			factor = w[i]
		}
		out[i] = v * scale * factor
	}
}

func (m *qwen) qkNorm(v, w []float32) {
	if w == nil {
		return
	}
	for o := 0; o < len(v); o += m.headSz {
		rmsnormInto(v[o:o+m.headSz], v[o:o+m.headSz], w, m.cfg.RMSEps)
	}
}

func (m *qwen) rope(h []float32, pos int, b *qblock) {
	half := m.headSz / 2
	for i := 0; i < half; i++ {
		freq := b.ropeFreq[i]
		angle := float64(pos) * freq
		sv, cv := math.Sincos(angle)
		s32, c32 := float32(sv), float32(cv)
		a, valB := h[i], h[i+half]
		h[i] = a*c32 - valB*s32
		h[i+half] = valB*c32 + a*s32
	}
}

func (m *qwen) attendHead(b *qblock, q, attn []float32, h, group, steps int, scores []float64) {
	qOff := h * m.headSz
	kvOff := (h / group) * m.headSz
	scale := 1.0 / math.Sqrt(float64(m.headSz))
	qh := q[qOff : qOff+m.headSz]
	maxs := math.Inf(-1)
	for t := 0; t < steps; t++ {
		var dot float64
		kh := b.kc[t][kvOff : kvOff+m.headSz]
		for i := 0; i < m.headSz; i++ {
			dot += float64(qh[i] * kh[i])
		}
		s := dot * scale
		scores[t] = s
		if s > maxs {
			maxs = s
		}
	}
	var sum float64
	for t := 0; t < steps; t++ {
		scores[t] = math.Exp(scores[t] - maxs)
		sum += scores[t]
	}
	inv := 1.0 / sum
	out := attn[qOff : qOff+m.headSz]
	for t := 0; t < steps; t++ {
		w := float32(scores[t] * inv)
		vh := b.vc[t][kvOff : kvOff+m.headSz]
		for i := 0; i < m.headSz; i++ {
			out[i] += w * vh[i]
		}
	}
}

func mv(x []float32, w *tensai.Matrix, q *qmat, bias []float32) []float32 {
	cols := 0
	if q != nil {
		cols = q.cols
	} else if w != nil {
		cols = w.Cols
	}
	out := make([]float32, cols)
	mvInto(out, x, w, q, bias)
	return out
}

func mvInto(out, x []float32, w *tensai.Matrix, q *qmat, bias []float32) {
	if q != nil {
		if err := q.f(x, out); err != nil {
			panic(err)
		}
		if bias != nil {
			for i, b := range bias {
				out[i] += b
			}
		}
		return
	}
	if w == nil {
		return
	}
	cols := w.Cols
	rows := min(w.Rows, len(x))
	if bias != nil {
		copy(out[:cols], bias[:cols])
	} else {
		for i := 0; i < cols; i++ {
			out[i] = 0
		}
	}
	for r := 0; r < rows; r++ {
		xr := x[r]
		if xr == 0 {
			continue
		}
		off := r * w.Cols
		for c := 0; c < cols; c++ {
			out[c] += xr * w.Data[off+c]
		}
	}
}

func (m *qwen) step(token, pos int) []float32 {
	cfg := m.cfg
	hs := cfg.HiddenSize
	group := cfg.Heads / cfg.KVHeads

	x := make([]float32, hs)
	copy(x, m.embed.Data[token*hs:(token+1)*hs])
	a := make([]float32, hs)

	kvDim := cfg.KVHeads * m.headSz
	qDim := cfg.Heads * m.headSz
	qkvW := qDim + 2*kvDim
	qkv := make([]float32, qkvW)
	attn := make([]float32, qDim)
	proj := make([]float32, hs)
	gu := make([]float32, 2*cfg.Intermediate)
	downBuf := make([]float32, hs)

	for li := range m.blocks {
		b := &m.blocks[li]
		rmsnormInto(a, x, b.ln1, cfg.RMSEps)
		mvInto(qkv, a, b.wQKV, b.qQKV, b.bQKV)
		q := qkv[:qDim]
		k := qkv[qDim : qDim+kvDim]
		v := qkv[qDim+kvDim:]
		m.qkNorm(q, b.qNorm)
		m.qkNorm(k, b.kNorm)
		if !b.noPE {
			for h := 0; h < cfg.Heads; h++ {
				m.rope(q[h*m.headSz:(h+1)*m.headSz], pos, b)
			}
			for h := 0; h < cfg.KVHeads; h++ {
				m.rope(k[h*m.headSz:(h+1)*m.headSz], pos, b)
			}
		}

		b.kc = append(b.kc, append(make([]float32, 0, kvDim), k...))
		b.vc = append(b.vc, append(make([]float32, 0, kvDim), v...))

		clear(attn)
		steps := len(b.kc)
		scores := make([]float64, steps)
		for h := 0; h < cfg.Heads; h++ {
			m.attendHead(b, q, attn, h, group, steps, scores)
		}
		mvInto(proj, attn, b.wo, b.qo, b.bo)
		if b.postAttn != nil {
			rmsnormInto(proj, proj, b.postAttn, cfg.RMSEps)
		}
		for i := range x {
			x[i] += proj[i]
		}

		rmsnormInto(a, x, b.ln2, cfg.RMSEps)
		mvInto(gu, a, b.wGU, b.qGU, nil)
		inter := cfg.Intermediate
		gate, up := gu[:inter], gu[inter:]
		for i := range gate {
			silu := gate[i] / (1.0 + float32(math.Exp(float64(-gate[i]))))
			gate[i] = silu * up[i]
		}
		mvInto(downBuf, gate, b.wDown, b.qDown, nil)
		if b.postFFN != nil {
			rmsnormInto(downBuf, downBuf, b.postFFN, cfg.RMSEps)
		}
		for i := range x {
			x[i] += downBuf[i]
		}
	}
	rmsnormInto(a, x, m.normW, cfg.RMSEps)
	return mv(a, m.lmT, m.qLmT, nil)
}
