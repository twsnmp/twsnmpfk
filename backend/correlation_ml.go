package backend

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	tensai "github.com/mattn/tensai"
	"github.com/mattn/tensai/autograd"
	"github.com/mattn/tensai/knn"
	"github.com/mattn/tensai/layer"
	"github.com/mattn/tensai/loss"
	"github.com/mattn/tensai/metrics"
	"github.com/mattn/tensai/model"
	"github.com/mattn/tensai/optim"
	"github.com/mattn/tensai/rnn"
)

const (
	mlMaxSamples      = 300
	mlTrainRatio      = 0.75
	mlpMinSamples     = 20
	lstmMinSamples    = 30
	knnMinSamples     = 15
	mlpEpochs         = 40
	lstmEpochs        = 25
	lstmHiddenSize    = 8
	lstmDefaultSeqLen = 6
)

type mlNormCols struct {
	names []string
	data  [][]float64 // [col][row]
}

// prepareMLSlice extracts a contiguous tail window (up to maxSamples) for temporal ML models.
func prepareMLSlice(names []string, cols [][]float64, maxSamples int) ([]string, [][]float64) {
	if len(cols) == 0 || len(cols[0]) == 0 {
		return nil, nil
	}
	n := len(cols[0])
	start := 0
	if maxSamples > 0 && n > maxSamples {
		start = n - maxSamples
	}
	outNames := make([]string, 0, len(cols))
	outCols := make([][]float64, 0, len(cols))
	for i, c := range cols {
		if len(c) != n {
			continue
		}
		name := fmt.Sprintf("v%d", i)
		if i < len(names) && names[i] != "" {
			name = names[i]
		}
		outNames = append(outNames, name)
		outCols = append(outCols, c[start:])
	}
	return outNames, outCols
}

// standardizeTrainTest standardizes columns using only [0:splitIdx] statistics to prevent data leakage.
// Zero-variance columns on the training split are dropped.
func standardizeTrainTest(names []string, cols [][]float64, splitIdx int) mlNormCols {
	if len(cols) == 0 || splitIdx < 2 {
		return mlNormCols{}
	}
	n := len(cols[0])
	res := mlNormCols{
		names: make([]string, 0, len(cols)),
		data:  make([][]float64, 0, len(cols)),
	}
	for cIdx, col := range cols {
		if len(col) != n {
			continue
		}
		var sum float64
		for i := 0; i < splitIdx; i++ {
			sum += sanitizeFloat(col[i])
		}
		mean := sum / float64(splitIdx)
		var sumSq float64
		for i := 0; i < splitIdx; i++ {
			d := sanitizeFloat(col[i]) - mean
			sumSq += d * d
		}
		sd := math.Sqrt(sumSq / float64(splitIdx))
		if sd < 1e-12 {
			continue
		}
		norm := make([]float64, n)
		for i := 0; i < n; i++ {
			norm[i] = sanitizeFloat((sanitizeFloat(col[i]) - mean) / sd)
		}
		res.names = append(res.names, names[cIdx])
		res.data = append(res.data, norm)
	}
	return res
}

func computeR2Score(yTrue, yPred []tensai.Float, rows, cols int) float64 {
	if rows < 2 || cols < 1 || len(yTrue) != rows*cols || len(yPred) != rows*cols {
		return 0
	}
	var sumR2 float64
	validCols := 0
	for c := 0; c < cols; c++ {
		var mean float64
		for r := 0; r < rows; r++ {
			mean += float64(yTrue[r*cols+c])
		}
		mean /= float64(rows)
		var ssTot, ssRes float64
		for r := 0; r < rows; r++ {
			yt := float64(yTrue[r*cols+c])
			yp := float64(yPred[r*cols+c])
			if math.IsNaN(yp) || math.IsInf(yp, 0) {
				return 0
			}
			dt := yt - mean
			dr := yt - yp
			ssTot += dt * dt
			ssRes += dr * dr
		}
		if ssTot < 1e-12 {
			continue
		}
		r2 := 1.0 - ssRes/ssTot
		sumR2 += r2
		validCols++
	}
	if validCols == 0 {
		return 0
	}
	avgR2 := sumR2 / float64(validCols)
	if avgR2 <= 0 {
		return 0
	}
	if avgR2 > 1.0 {
		avgR2 = 1.0
	}
	return sanitizeFloat(math.Sqrt(avgR2))
}

func computeMSE(yTrue, yPred []tensai.Float) float64 {
	if len(yTrue) == 0 || len(yTrue) != len(yPred) {
		return 0
	}
	var sum float64
	for i := range yTrue {
		yp := float64(yPred[i])
		if math.IsNaN(yp) || math.IsInf(yp, 0) {
			return math.Inf(1)
		}
		d := float64(yTrue[i]) - yp
		sum += d * d
	}
	return sum / float64(len(yTrue))
}

type mlpFeatureSpec struct {
	name string
	varN string
	lag  int
	col  int
}

func evalMLPDirection(xNorm, yNorm mlNormCols, lagWindow int) (float64, []CorrelationFeatureImportance) {
	if len(xNorm.data) == 0 || len(yNorm.data) == 0 {
		return 0, nil
	}
	n := len(xNorm.data[0])
	if lagWindow < 0 {
		lagWindow = 0
	}
	if n-lagWindow < mlpMinSamples {
		lagWindow = 0
	}
	effN := n - lagWindow
	trainRows := int(float64(effN) * mlTrainRatio)
	testRows := effN - trainRows
	if trainRows < 10 || testRows < 5 {
		return 0, nil
	}

	inCols := len(xNorm.data) * (lagWindow + 1)
	outCols := len(yNorm.data)

	specs := make([]mlpFeatureSpec, 0, inCols)
	for cIdx, vName := range xNorm.names {
		for l := 0; l <= lagWindow; l++ {
			featName := vName
			if l > 0 {
				featName = fmt.Sprintf("%s (t-%d)", vName, l)
			}
			specs = append(specs, mlpFeatureSpec{
				name: featName,
				varN: vName,
				lag:  l,
				col:  cIdx,
			})
		}
	}

	buildIO := func(startSample, count int) ([]tensai.Float, []tensai.Float) {
		xBuf := make([]tensai.Float, count*inCols)
		yBuf := make([]tensai.Float, count*outCols)
		for i := 0; i < count; i++ {
			t := lagWindow + startSample + i
			for fIdx, sp := range specs {
				xBuf[i*inCols+fIdx] = tensai.Float(xNorm.data[sp.col][t-sp.lag])
			}
			for oc := 0; oc < outCols; oc++ {
				yBuf[i*outCols+oc] = tensai.Float(yNorm.data[oc][t])
			}
		}
		return xBuf, yBuf
	}

	trX, trY := buildIO(0, trainRows)
	teX, teY := buildIO(trainRows, testRows)

	trXMat, err := tensai.NewMatrixFromSlice(trainRows, inCols, trX)
	if err != nil {
		return 0, nil
	}
	trYMat, err := tensai.NewMatrixFromSlice(trainRows, outCols, trY)
	if err != nil {
		return 0, nil
	}
	teXMat, err := tensai.NewMatrixFromSlice(testRows, inCols, teX)
	if err != nil {
		return 0, nil
	}

	hidden := inCols * 2
	if hidden < 8 {
		hidden = 8
	}
	if hidden > 16 {
		hidden = 16
	}

	net := model.NewSequential()
	net.Add(layer.NewDense(hidden))
	net.Add(&layer.Tanh{})
	net.Add(layer.NewDense(outCols))

	if err := net.Compile(inCols, loss.MeanSquaredError{}, optim.NewAdamW(0.02, 1e-4)); err != nil {
		return 0, nil
	}
	for e := 0; e < mlpEpochs; e++ {
		if _, err := net.FitStep(trXMat, trYMat); err != nil {
			return 0, nil
		}
	}

	predMat, err := net.Predict(teXMat)
	if err != nil {
		return 0, nil
	}
	score := computeR2Score(teY, predMat.Data, testRows, outCols)
	if score <= 0 {
		return 0, nil
	}

	// Permutation importance on test set
	baseMSE := computeMSE(teY, predMat.Data)
	if math.IsInf(baseMSE, 1) || math.IsNaN(baseMSE) {
		return score, nil
	}

	rng := rand.New(rand.NewPCG(42, 1))
	rawImps := make([]float64, inCols)
	var sumImp float64
	permBuf := make([]tensai.Float, len(teX))
	colVals := make([]tensai.Float, testRows)
	permIdx := make([]int, testRows)
	for i := range permIdx {
		permIdx[i] = i
	}

	for fIdx := 0; fIdx < inCols; fIdx++ {
		copy(permBuf, teX)
		for r := 0; r < testRows; r++ {
			colVals[r] = teX[r*inCols+fIdx]
		}
		rng.Shuffle(testRows, func(i, j int) {
			permIdx[i], permIdx[j] = permIdx[j], permIdx[i]
		})
		for r := 0; r < testRows; r++ {
			permBuf[r*inCols+fIdx] = colVals[permIdx[r]]
		}
		pMat, err := tensai.NewMatrixFromSlice(testRows, inCols, permBuf)
		if err != nil {
			continue
		}
		pPred, err := net.Predict(pMat)
		if err != nil {
			continue
		}
		pMSE := computeMSE(teY, pPred.Data)
		diff := pMSE - baseMSE
		if diff > 0 && !math.IsInf(diff, 1) && !math.IsNaN(diff) {
			rawImps[fIdx] = diff
			sumImp += diff
		}
	}

	imps := make([]CorrelationFeatureImportance, 0, inCols)
	for fIdx, sp := range specs {
		normImp := 0.0
		if sumImp > 1e-12 {
			normImp = rawImps[fIdx] / sumImp
		}
		if normImp > 0 {
			imps = append(imps, CorrelationFeatureImportance{
				Feature:    sp.name,
				Var:        sp.varN,
				Lag:        sp.lag,
				Importance: sanitizeFloat(normImp),
			})
		}
	}
	sort.Slice(imps, func(i, j int) bool {
		return imps[i].Importance > imps[j].Importance
	})
	if len(imps) > 10 {
		imps = imps[:10]
	}

	return score, imps
}

// calcMLPCorrelation computes bidirectional nonlinear MLP predictability (with lag window) and permutation importance.
func calcMLPCorrelation(names1 []string, cols1 [][]float64, names2 []string, cols2 [][]float64, maxLag int) (score, pValue float64, imps []CorrelationFeatureImportance) {
	n1, c1 := prepareMLSlice(names1, cols1, mlMaxSamples)
	n2, c2 := prepareMLSlice(names2, cols2, mlMaxSamples)
	if len(c1) == 0 || len(c2) == 0 {
		return 0, 1.0, nil
	}
	n := len(c1[0])
	if n < mlpMinSamples || len(c2[0]) != n {
		return 0, 1.0, nil
	}

	lagWindow := 2
	if maxLag > 0 && maxLag < lagWindow {
		lagWindow = maxLag
	}
	if n >= 60 && maxLag >= 3 {
		lagWindow = 3
	}

	splitIdx := int(float64(n) * mlTrainRatio)
	xNorm := standardizeTrainTest(n1, c1, splitIdx)
	yNorm := standardizeTrainTest(n2, c2, splitIdx)
	if len(xNorm.data) == 0 || len(yNorm.data) == 0 {
		return 0, 1.0, nil
	}

	sXY, impsXY := evalMLPDirection(xNorm, yNorm, lagWindow)
	sYX, impsYX := evalMLPDirection(yNorm, xNorm, lagWindow)

	if sXY >= sYX {
		return sXY, 1.0, impsXY
	}
	return sYX, 1.0, impsYX
}

func evalLSTMDirection(xNorm, yNorm mlNormCols, seqLen int) float64 {
	if len(xNorm.data) == 0 || len(yNorm.data) == 0 {
		return 0
	}
	n := len(xNorm.data[0])
	if seqLen < 2 || n-seqLen < 15 {
		return 0
	}
	effN := n - seqLen + 1
	trainRows := int(float64(effN) * mlTrainRatio)
	testRows := effN - trainRows
	if trainRows < 10 || testRows < 5 {
		return 0
	}

	inDim := len(xNorm.data)
	outDim := len(yNorm.data)

	buildSeq := func(startSample, count int) ([]*tensai.Matrix, []tensai.Float) {
		steps := make([]*tensai.Matrix, seqLen)
		for s := 0; s < seqLen; s++ {
			buf := make([]tensai.Float, count*inDim)
			for i := 0; i < count; i++ {
				t := startSample + i + s
				for c := 0; c < inDim; c++ {
					buf[i*inDim+c] = tensai.Float(xNorm.data[c][t])
				}
			}
			m, _ := tensai.NewMatrixFromSlice(count, inDim, buf)
			steps[s] = m
		}
		yBuf := make([]tensai.Float, count*outDim)
		for i := 0; i < count; i++ {
			t := startSample + i + seqLen - 1
			for c := 0; c < outDim; c++ {
				yBuf[i*outDim+c] = tensai.Float(yNorm.data[c][t])
			}
		}
		return steps, yBuf
	}

	trSteps, trY := buildSeq(0, trainRows)
	teSteps, teY := buildSeq(trainRows, testRows)
	trTgt, err := tensai.NewTensorFromSlice(trY, trainRows, outDim)
	if err != nil {
		return 0
	}

	rng := rand.New(rand.NewPCG(42, 1))
	cell := rnn.NewLSTMCell(inDim, lstmHiddenSize, rng)
	wOut := autograd.Param(tensai.RandomMatrix(lstmHiddenSize, outDim, rng))
	bOut := autograd.Param(tensai.NewMatrix(1, outDim))

	params := append(cell.Params(), wOut, bOut)
	tape := autograd.NewTape()
	tape.Bind(params...)
	trainer := autograd.NewTrainer(optim.NewAdamW(0.03, 1e-4), params...)

	for e := 0; e < lstmEpochs; e++ {
		h, c := cell.InitState(trainRows)
		for s := 0; s < seqLen; s++ {
			h, c = cell.Step(autograd.Input(trSteps[s]), h, c)
		}
		pred := h.MatMul(wOut).AddRow(bOut)
		lossNode := pred.MSELoss(trTgt)
		lVal := trainer.Step(lossNode)
		tape.Reset()
		if math.IsNaN(float64(lVal)) || math.IsInf(float64(lVal), 0) {
			return 0
		}
	}

	// Evaluate on test split
	h, c := cell.InitState(testRows)
	for s := 0; s < seqLen; s++ {
		h, c = cell.Step(autograd.Input(teSteps[s]), h, c)
	}
	pred := h.MatMul(wOut).AddRow(bOut)
	predData := append([]tensai.Float(nil), pred.Value().Data...)
	tape.Reset()

	return computeR2Score(teY, predData, testRows, outDim)
}

// calcLSTMCorrelation computes bidirectional LSTM sequential predictability using tensai/rnn and autograd.
func calcLSTMCorrelation(names1 []string, cols1 [][]float64, names2 []string, cols2 [][]float64) (score, pValue float64) {
	n1, c1 := prepareMLSlice(names1, cols1, mlMaxSamples)
	n2, c2 := prepareMLSlice(names2, cols2, mlMaxSamples)
	if len(c1) == 0 || len(c2) == 0 {
		return 0, 1.0
	}
	n := len(c1[0])
	if n < lstmMinSamples || len(c2[0]) != n {
		return 0, 1.0
	}

	seqLen := lstmDefaultSeqLen
	if n < 45 {
		seqLen = 4
	}

	splitIdx := int(float64(n) * mlTrainRatio)
	xNorm := standardizeTrainTest(n1, c1, splitIdx)
	yNorm := standardizeTrainTest(n2, c2, splitIdx)
	if len(xNorm.data) == 0 || len(yNorm.data) == 0 {
		return 0, 1.0
	}

	sXY := evalLSTMDirection(xNorm, yNorm, seqLen)
	sYX := evalLSTMDirection(yNorm, xNorm, seqLen)
	if sXY >= sYX {
		return sXY, 1.0
	}
	return sYX, 1.0
}

func selectPrimaryTargetCol(norm mlNormCols) (int, string) {
	if len(norm.data) == 0 {
		return -1, ""
	}
	return 0, norm.names[0]
}

func evalKNNDirection(xNorm, yNorm mlNormCols) (float64, *CorrelationRegimeResult) {
	if len(xNorm.data) == 0 || len(yNorm.data) == 0 {
		return 0, nil
	}
	n := len(xNorm.data[0])
	if n < knnMinSamples {
		return 0, nil
	}
	trainRows := int(float64(n) * mlTrainRatio)
	testRows := n - trainRows
	if trainRows < 9 || testRows < 4 {
		return 0, nil
	}

	targetIdx, targetName := selectPrimaryTargetCol(yNorm)
	if targetIdx < 0 {
		return 0, nil
	}
	targetCol := yNorm.data[targetIdx]

	// Determine quantile thresholds from training slice only
	trainSorted := make([]float64, trainRows)
	copy(trainSorted, targetCol[:trainRows])
	sort.Float64s(trainSorted)

	q1 := trainSorted[trainRows/3]
	q2 := trainSorted[(2*trainRows)/3]

	var labels []string
	classify := func(v float64) int {
		if q2-q1 < 1e-9 {
			// Binary regime fallback
			if v <= q1 {
				return 0
			}
			return 1
		}
		if v <= q1 {
			return 0
		}
		if v <= q2 {
			return 1
		}
		return 2
	}
	if q2-q1 < 1e-9 {
		if trainSorted[trainRows-1]-trainSorted[0] < 1e-9 {
			return 0, nil
		}
		labels = []string{"Low", "High"}
	} else {
		labels = []string{"Low", "Normal", "High"}
	}

	inCols := len(xNorm.data)
	trX := make([]tensai.Float, trainRows*inCols)
	trY := make([]tensai.Float, trainRows)
	for r := 0; r < trainRows; r++ {
		for c := 0; c < inCols; c++ {
			trX[r*inCols+c] = tensai.Float(xNorm.data[c][r])
		}
		trY[r] = tensai.Float(classify(targetCol[r]))
	}

	teX := make([]tensai.Float, testRows*inCols)
	teY := make([]tensai.Float, testRows)
	for r := 0; r < testRows; r++ {
		idx := trainRows + r
		for c := 0; c < inCols; c++ {
			teX[r*inCols+c] = tensai.Float(xNorm.data[c][idx])
		}
		teY[r] = tensai.Float(classify(targetCol[idx]))
	}

	trXMat, err := tensai.NewMatrixFromSlice(trainRows, inCols, trX)
	if err != nil {
		return 0, nil
	}
	trYMat, err := tensai.NewMatrixFromSlice(trainRows, 1, trY)
	if err != nil {
		return 0, nil
	}
	teXMat, err := tensai.NewMatrixFromSlice(testRows, inCols, teX)
	if err != nil {
		return 0, nil
	}
	teYMat, err := tensai.NewMatrixFromSlice(testRows, 1, teY)
	if err != nil {
		return 0, nil
	}

	k := 5
	if trainRows < 15 {
		k = 3
	}
	clf := knn.New(k)
	if err := clf.Fit(trXMat, trYMat); err != nil {
		return 0, nil
	}
	predMat, err := clf.Predict(teXMat)
	if err != nil {
		return 0, nil
	}
	rep, err := metrics.Report(predMat, teYMat)
	if err != nil {
		return 0, nil
	}

	numClasses := len(labels)
	conf := make([][]int, numClasses)
	for i := range conf {
		conf[i] = make([]int, numClasses)
	}
	for r := 0; r < len(rep.Confusion) && r < numClasses; r++ {
		for c := 0; c < len(rep.Confusion[r]) && c < numClasses; c++ {
			conf[r][c] = rep.Confusion[r][c]
		}
	}

	// Compute Balanced Accuracy to guard against class imbalance
	var sumRecall float64
	activeClasses := 0
	for c := 0; c < numClasses; c++ {
		var rowSum int
		for p := 0; p < numClasses; p++ {
			rowSum += conf[c][p]
		}
		if rowSum > 0 {
			sumRecall += float64(conf[c][c]) / float64(rowSum)
			activeClasses++
		}
	}
	if activeClasses < 2 {
		return 0, nil
	}
	balAcc := sumRecall / float64(activeClasses)
	chance := 1.0 / float64(activeClasses)
	score := (balAcc - chance) / (1.0 - chance)
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return sanitizeFloat(score), &CorrelationRegimeResult{
		Accuracy:         sanitizeFloat(rep.Accuracy()),
		BalancedAccuracy: sanitizeFloat(balAcc),
		Labels:           labels,
		Confusion:        conf,
		TargetVar:        targetName,
	}
}

// calcKNNRegimeCorrelation evaluates how well X patterns classify Y into Low/Normal/High regimes using tensai/knn.
func calcKNNRegimeCorrelation(names1 []string, cols1 [][]float64, names2 []string, cols2 [][]float64) (score, pValue float64, regime *CorrelationRegimeResult) {
	n1, c1 := prepareMLSlice(names1, cols1, mlMaxSamples)
	n2, c2 := prepareMLSlice(names2, cols2, mlMaxSamples)
	if len(c1) == 0 || len(c2) == 0 {
		return 0, 1.0, nil
	}
	n := len(c1[0])
	if n < knnMinSamples || len(c2[0]) != n {
		return 0, 1.0, nil
	}

	splitIdx := int(float64(n) * mlTrainRatio)
	xNorm := standardizeTrainTest(n1, c1, splitIdx)
	yNorm := standardizeTrainTest(n2, c2, splitIdx)
	if len(xNorm.data) == 0 || len(yNorm.data) == 0 {
		return 0, 1.0, nil
	}

	sXY, regXY := evalKNNDirection(xNorm, yNorm)
	sYX, regYX := evalKNNDirection(yNorm, xNorm)
	if sXY >= sYX {
		return sXY, 1.0, regXY
	}
	return sYX, 1.0, regYX
}
