package background

type Worker struct {
	running bool
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Start() {
	w.running = true
}

func (w *Worker) Stop() {
	w.running = false
}

func (w *Worker) IsRunning() bool {
	return w.running
}

func (w *Worker) Status() string {
	if w.running {
		return "running"
	}
	return "stopped"
}
