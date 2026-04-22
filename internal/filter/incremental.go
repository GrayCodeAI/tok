package filter

// IncrementalCompressor compresses data incrementally
type IncrementalCompressor struct {
	coordinator *PipelineCoordinator
	buffer      string
	chunkSize   int
}

func NewIncrementalCompressor(cfg PipelineConfig, chunkSize int) *IncrementalCompressor {
	return &IncrementalCompressor{
		coordinator: NewPipelineCoordinator(&cfg),
		chunkSize:   chunkSize,
	}
}

func (ic *IncrementalCompressor) Add(data string) string {
	ic.buffer += data

	if len(ic.buffer) < ic.chunkSize {
		return ""
	}

	output, _ := ic.coordinator.Process(ic.buffer)
	ic.buffer = ""
	return output
}

func (ic *IncrementalCompressor) Flush() string {
	if len(ic.buffer) == 0 {
		return ""
	}

	output, _ := ic.coordinator.Process(ic.buffer)
	ic.buffer = ""
	return output
}
