package githubembed

import "fmt"

type EmbedType string

const (
	EmbedBadge EmbedType = "badge"
	EmbedCard  EmbedType = "card"
	Embed3D    EmbedType = "3d"
)

type GitHubEmbed struct {
	Type    EmbedType `json:"type"`
	Content string    `json:"content"`
	URL     string    `json:"url"`
}

type EmbedEngine struct {
	badges map[string]*GitHubEmbed
}

func NewEmbedEngine() *EmbedEngine {
	return &EmbedEngine{
		badges: make(map[string]*GitHubEmbed),
	}
}

func (e *EmbedEngine) GenerateBadge(username string, savings int64, tier string) *GitHubEmbed {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="30">
<rect width="200" height="30" rx="5" fill="#1a1a2e"/>
<text x="10" y="20" fill="#e94560" font-size="11" font-weight="bold">TokMan</text>
<text x="70" y="20" fill="#eee" font-size="11">%s tokens saved</text>
</svg>`, formatSavings(savings))

	embed := &GitHubEmbed{
		Type:    EmbedBadge,
		Content: svg,
	}
	e.badges[username] = embed
	return embed
}

func (e *EmbedEngine) GenerateCard(username string, savings int64, tier string, commands int) *GitHubEmbed {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="400" height="200">
<rect width="400" height="200" rx="10" fill="#1a1a2e"/>
<text x="20" y="35" fill="#e94560" font-size="20" font-weight="bold">TokMan</text>
<text x="20" y="70" fill="#eee" font-size="14">User: %s</text>
<text x="20" y="100" fill="#eee" font-size="14">Tokens Saved: %s</text>
<text x="20" y="130" fill="#eee" font-size="14">Tier: %s</text>
<text x="20" y="160" fill="#eee" font-size="14">Commands: %d</text>
<text x="200" y="35" fill="#16c79a" font-size="14">Compression Stats</text>
</svg>`, username, formatSavings(savings), tier, commands)

	return &GitHubEmbed{
		Type:    EmbedCard,
		Content: svg,
	}
}

func (e *EmbedEngine) Generate3DGraph(username string, days []int64) *GitHubEmbed {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="200">
<rect width="300" height="200" rx="5" fill="#1a1a2e"/>
`
	for i, val := range days {
		if i >= 30 {
			break
		}
		height := int(val / 100)
		if height > 100 {
			height = 100
		}
		if height < 1 {
			height = 1
		}
		x := 20 + i*9
		y := 180 - height
		svg += fmt.Sprintf(`<rect x="%d" y="%d" width="7" height="%d" fill="#e94560" rx="1"/>`, x, y, height)
	}
	svg += "</svg>"

	return &GitHubEmbed{
		Type:    Embed3D,
		Content: svg,
	}
}

func (e *EmbedEngine) GetBadge(username string) *GitHubEmbed {
	return e.badges[username]
}

func (e *EmbedEngine) EmbedMarkdown(username string, embed *GitHubEmbed) string {
	return fmt.Sprintf("![TokMan %s](%s)", username, embed.URL)
}

func formatSavings(n int64) string {
	if n >= 1000000000 {
		return fmt.Sprintf("%dB", n/1000000000)
	}
	if n >= 1000000 {
		return fmt.Sprintf("%dM", n/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dK", n/1000)
	}
	return fmt.Sprintf("%d", n)
}
