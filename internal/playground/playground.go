package playground

type PlaygroundSession struct {
	ID           string              `json:"id"`
	Model        string              `json:"model"`
	Temperature  float64             `json:"temperature"`
	MaxTokens    int                 `json:"max_tokens"`
	History      []PlaygroundMessage `json:"history"`
	CostEstimate float64             `json:"cost_estimate"`
	Source       string              `json:"source"`
}

type PlaygroundMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}

type Playground struct {
	sessions map[string]*PlaygroundSession
}

func NewPlayground() *Playground {
	return &Playground{
		sessions: make(map[string]*PlaygroundSession),
	}
}

func (p *Playground) CreateSession(id, model string) *PlaygroundSession {
	session := &PlaygroundSession{
		ID:          id,
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   4096,
	}
	p.sessions[id] = session
	return session
}

func (p *Playground) GetSession(id string) *PlaygroundSession {
	return p.sessions[id]
}

func (p *Playground) AddMessage(sessionID, role, content string, tokens int) {
	if session, ok := p.sessions[sessionID]; ok {
		session.History = append(session.History, PlaygroundMessage{
			Role:    role,
			Content: content,
			Tokens:  tokens,
		})
	}
}

func (p *Playground) EstimateCost(sessionID string) float64 {
	session := p.sessions[sessionID]
	if session == nil {
		return 0
	}
	totalTokens := 0
	for _, msg := range session.History {
		totalTokens += msg.Tokens
	}
	return float64(totalTokens) * 0.00001
}

func (p *Playground) ListSessions() []*PlaygroundSession {
	var sessions []*PlaygroundSession
	for _, s := range p.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}
