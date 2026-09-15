package tensai

import (
	"encoding/binary"
	"fmt"

	tensai "github.com/mattn/tensai"
	"github.com/mattn/tensai/encoding/gguf"
	"github.com/mattn/tensai/quant"
	"github.com/mattn/tensai/tokenizer"
)

func repackQ8(dst *quant.Q8GMatrix, raw []byte, out, in, colOff int) {
	nb := in / 32
	for r := 0; r < out; r++ {
		j := colOff + r
		for b := 0; b < nb; b++ {
			blk := raw[(r*nb+b)*34:]
			dst.Scale[dst.TableIndex(b, j)] = gguf.Float16(binary.LittleEndian.Uint16(blk))
			var sum int32
			for i := 0; i < 32; i++ {
				w := int8(blk[2+i])
				dst.Q[dst.Index(b*32+i, j)] = w
				sum += int32(w)
			}
			dst.ColSum64[dst.TableIndex(b, j)] = 64 * sum
		}
	}
}

func loadQwenGGUF(path string, bits int, forGPU bool) (*qwen, *tokenizer.Tokenizer, error) {
	g, err := gguf.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer g.Close()

	arch, _ := g.String("general.architecture")
	if arch == "" {
		arch = "qwen2"
	}

	meta := func(key string) int64 {
		n, _ := g.Int(arch + "." + key)
		return n
	}

	var cfg config
	cfg.ModelType = arch
	cfg.HiddenSize = int(meta("embedding_length"))
	cfg.Intermediate = int(meta("feed_forward_length"))
	cfg.Layers = int(meta("block_count"))
	cfg.Heads = int(meta("attention.head_count"))
	cfg.KVHeads = int(meta("attention.head_count_kv"))
	cfg.MaxPos = int(meta("context_length"))
	cfg.Vocab = int(meta("vocab_size"))
	cfg.HeadDim = int(meta("attention.key_length"))
	cfg.RMSEps, _ = g.Float(arch + ".attention.layer_norm_rms_epsilon")
	cfg.RopeTheta, _ = g.Float(arch + ".rope.freq_base")
	if cfg.RopeTheta == 0 {
		cfg.RopeTheta = 1000000.0
	}
	if cfg.RMSEps == 0 {
		cfg.RMSEps = 1e-6
	}

	tok, err := buildGGUFTokenizer(g)
	if err != nil {
		return nil, nil, err
	}

	tensor := func(name string) *tensai.Tensor {
		t, err := g.Tensor(name)
		if err != nil {
			panic(err)
		}
		return t
	}
	vecOpt := func(name string) []float32 {
		t, err := g.Tensor(name)
		if err != nil {
			return nil
		}
		return t.Data
	}

	headSz := cfg.HiddenSize / cfg.Heads
	if cfg.HeadDim != 0 {
		headSz = cfg.HeadDim
	}

	m := &qwen{cfg: cfg, headSz: headSz}
	m.embed = tensor("token_embd.weight")
	m.normW = tensor("output_norm.weight").Data
	m.blocks = make([]qblock, cfg.Layers)

	allQ8 := func(names ...string) bool {
		if forGPU || bits != 8 {
			return false
		}
		for _, name := range names {
			typ, shape, ok := g.Info(name)
			if !ok || typ != "Q8_0" || shape[1]%32 != 0 {
				return false
			}
		}
		return true
	}

	linDirect := func(names []string) *qmat {
		var outs []int
		var in int
		for _, name := range names {
			_, shape, _ := g.Info(name)
			outs = append(outs, shape[0])
			in = shape[1]
		}
		total := 0
		for _, o := range outs {
			total += o
		}
		dst := quant.NewQ8GMatrix(in, total, 0)
		colOff := 0
		for i, name := range names {
			_, raw, err := g.RawTensor(name)
			if err != nil {
				panic(err)
			}
			repackQ8(dst, raw, outs[i], in, colOff)
			colOff += outs[i]
		}
		return qmatQ8G(dst)
	}

	trans := func(name string) *tensai.Matrix {
		m, err := tensor(name).Matrix()
		if err != nil {
			panic(err)
		}
		return m.T()
	}

	lin := func(names ...string) (*tensai.Matrix, *qmat) {
		if allQ8(names...) {
			return nil, linDirect(names)
		}
		var parts []*tensai.Matrix
		for _, name := range names {
			parts = append(parts, trans(name))
		}
		var w *tensai.Matrix
		if len(parts) == 1 {
			w = parts[0]
		} else {
			rows := parts[0].Rows
			cols := 0
			for _, p := range parts {
				cols += p.Cols
			}
			w = tensai.NewMatrix(rows, cols)
			for r := 0; r < rows; r++ {
				cOff := 0
				for _, p := range parts {
					copy(w.Data[r*cols+cOff:r*cols+cOff+p.Cols], p.Data[r*p.Cols:(r+1)*p.Cols])
					cOff += p.Cols
				}
			}
		}
		if bits == 0 {
			return w, nil
		}
		return nil, quantizeMat(w, bits)
	}

	// LM Head
	if _, _, ok := g.Info("output.weight"); ok {
		m.lmT, m.qLmT = lin("output.weight")
	} else {
		em, err := m.embed.Matrix()
		if err != nil {
			panic(err)
		}
		lmT := em.T()
		if bits == 0 {
			m.lmT = lmT
		} else {
			m.qLmT = quantizeMat(lmT, bits)
		}
	}

	catVec := func(vecs ...[]float32) []float32 {
		var n int
		hasAny := false
		for _, v := range vecs {
			if v != nil {
				hasAny = true
				n += len(v)
			}
		}
		if !hasAny {
			return nil
		}
		out := make([]float32, n)
		off := 0
		for _, v := range vecs {
			if v != nil {
				copy(out[off:off+len(v)], v)
				off += len(v)
			}
		}
		return out
	}

	for i := range m.blocks {
		b := &m.blocks[i]
		p := fmt.Sprintf("blk.%d.", i)
		b.ln1 = tensor(p + "attn_norm.weight").Data
		b.ln2 = tensor(p + "ffn_norm.weight").Data
		b.qNorm = vecOpt(p + "attn_q_norm.weight")
		b.kNorm = vecOpt(p + "attn_k_norm.weight")

		b.wQKV, b.qQKV = lin(p+"attn_q.weight", p+"attn_k.weight", p+"attn_v.weight")
		b.wo, b.qo = lin(p + "attn_output.weight")
		b.wGU, b.qGU = lin(p+"ffn_gate.weight", p+"ffn_up.weight")
		b.wDown, b.qDown = lin(p + "ffn_down.weight")

		b.bQKV = catVec(
			vecOpt(p+"attn_q.bias"),
			vecOpt(p+"attn_k.bias"),
			vecOpt(p+"attn_v.bias"),
		)
		b.bo = vecOpt(p + "attn_output.bias")
	}

	m.initRopeFreqs()
	return m, tok, nil
}
