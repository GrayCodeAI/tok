package discover

type Discover struct{}

func New() *Discover {
	return &Discover{}
}

func RewriteCommand(cmd string, args []string) (string, []string) {
	return cmd, args
}
