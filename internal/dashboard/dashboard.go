package dashboard

type Dashboard struct{}

func New() *Dashboard {
	return &Dashboard{}
}

type Cmd struct{}

func NewCmd() *Cmd {
	return &Cmd{}
}

func (c *Cmd) Run() error {
	return nil
}
