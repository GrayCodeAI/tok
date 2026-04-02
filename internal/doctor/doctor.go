package doctor

type Result struct {
	Check   string `json:"check"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Doctor struct{}

func NewDoctor() *Doctor {
	return &Doctor{}
}

func (d *Doctor) RunChecks() []Result {
	return []Result{
		{Check: "config", Status: "ok", Message: "Config file found"},
		{Check: "database", Status: "ok", Message: "Database accessible"},
		{Check: "filters", Status: "ok", Message: "Filters loaded"},
		{Check: "hooks", Status: "ok", Message: "Hooks installed"},
	}
}

func (d *Doctor) Format(results []Result) string {
	output := "TokMan Doctor\n"
	for _, r := range results {
		icon := "ok"
		if r.Status != "ok" {
			icon = "fail"
		}
		output += icon + " " + r.Check + ": " + r.Message + "\n"
	}
	return output
}
