package shieldsbadge

type Badge struct {
	Subject string `json:"subject"`
	Status  string `json:"status"`
	Color   string `json:"color"`
}

func NewBadge(subject, status, color string) *Badge {
	return &Badge{Subject: subject, Status: status, Color: color}
}

func (b *Badge) SVG() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" width="120" height="20">
<rect width="120" height="20" rx="3" fill="#` + b.Color + `"/>
<text x="60" y="14" text-anchor="middle" fill="#fff" font-size="12" font-weight="bold">` + b.Subject + ` ` + b.Status + `</text>
</svg>`
}
