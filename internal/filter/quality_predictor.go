package filter

// QualityPredictor predicts compression quality before processing
type QualityPredictor struct {
	history map[string]float64
}

func NewQualityPredictor() *QualityPredictor {
	return &QualityPredictor{history: make(map[string]float64)}
}

func (qp *QualityPredictor) Predict(input string) float64 {
	features := qp.extractFeatures(input)
	key := features.String()

	if score, ok := qp.history[key]; ok {
		return score
	}

	return 0.7 // Default prediction
}

func (qp *QualityPredictor) Learn(input string, actualRatio float64) {
	features := qp.extractFeatures(input)
	qp.history[features.String()] = actualRatio
}

type Features struct {
	Length      int
	Entropy     float64
	ContentType string
}

func (f Features) String() string {
	return f.ContentType
}

func (qp *QualityPredictor) extractFeatures(input string) Features {
	return Features{
		Length:      len(input),
		Entropy:     0.5,
		ContentType: detectContentType(input),
	}
}
