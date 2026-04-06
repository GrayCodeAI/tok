package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// allowedOrigins lists origins permitted for WebSocket connections.
var allowedOrigins = []string{"localhost", "127.0.0.1"}

// SetAllowedOrigins configures which origins are allowed for WebSocket.
func SetAllowedOrigins(origins []string) {
	allowedOrigins = origins
}

// LiveServer provides WebSocket-based live updates.
type LiveServer struct {
	upgrader     websocket.Upgrader
	clients      map[*Client]bool
	broadcast    chan Message
	register     chan *Client
	unregister   chan *Client
	mu           sync.RWMutex
	onConnect    func(*Client)
	onDisconnect func(*Client)
}

// Client represents a WebSocket client.
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	server *LiveServer
	id     string
}

// Message represents a WebSocket message.
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func isAllowedOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // Same-origin requests have no Origin header
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, allowed := range allowedOrigins {
		if host == allowed || host == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// NewLiveServer creates a new live server.
func NewLiveServer() *LiveServer {
	return &LiveServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: isAllowedOrigin,
		},
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Start starts the live server.
func (s *LiveServer) Start() {
	go s.run()
}

// run is the main event loop.
func (s *LiveServer) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			if s.onConnect != nil {
				s.onConnect(client)
			}

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mu.Unlock()
			if s.onDisconnect != nil {
				s.onDisconnect(client)
			}

		case message := <-s.broadcast:
			data, err := json.Marshal(message)
			if err != nil {
				continue
			}
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.send <- data:
				default:
					// Client buffer full, close connection
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.RUnlock()

		case <-ticker.C:
			// Send ping to keep connections alive
			s.Broadcast("ping", map[string]interface{}{
				"time": time.Now().Unix(),
			})
		}
	}
}

// HandleWebSocket handles WebSocket connections.
func (s *LiveServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		server: s,
		id:     generateClientID(),
	}

	s.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// Broadcast sends a message to all clients.
func (s *LiveServer) Broadcast(msgType string, payload interface{}) {
	s.broadcast <- Message{
		Type:    msgType,
		Payload: payload,
	}
}

// BroadcastTo sends a message to a specific client.
func (s *LiveServer) BroadcastTo(clientID string, msgType string, payload interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.Marshal(Message{
		Type:    msgType,
		Payload: payload,
	})
	if err != nil {
		return
	}

	for client := range s.clients {
		if client.id == clientID {
			select {
			case client.send <- data:
			default:
			}
			return
		}
	}
}

// ClientCount returns the number of connected clients.
func (s *LiveServer) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// Client methods

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// LiveTicker provides token savings ticker updates.
type LiveTicker struct {
	server   *LiveServer
	tracker  *TokenTracker
	stop     chan bool
	interval time.Duration
}

// TokenTracker tracks token savings.
type TokenTracker struct {
	mu           sync.RWMutex
	totalSaved   int64
	totalInput   int64
	totalOutput  int64
	sessionStart time.Time
	history      []DataPoint
}

// DataPoint represents a data point in time.
type DataPoint struct {
	Timestamp int64 `json:"timestamp"`
	Saved     int64 `json:"saved"`
	Input     int64 `json:"input"`
	Output    int64 `json:"output"`
}

// NewTokenTracker creates a new token tracker.
func NewTokenTracker() *TokenTracker {
	return &TokenTracker{
		sessionStart: time.Now(),
		history:      make([]DataPoint, 0, 1000),
	}
}

// Record records token savings.
func (t *TokenTracker) Record(input, output int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	saved := input - output
	if saved < 0 {
		saved = 0
	}

	t.totalInput += int64(input)
	t.totalOutput += int64(output)
	t.totalSaved += int64(saved)

	// Add to history every 10 seconds
	if len(t.history) == 0 || time.Now().Unix()-t.history[len(t.history)-1].Timestamp >= 10 {
		t.history = append(t.history, DataPoint{
			Timestamp: time.Now().Unix(),
			Saved:     t.totalSaved,
			Input:     t.totalInput,
			Output:    t.totalOutput,
		})

		// Keep last 1000 points
		if len(t.history) > 1000 {
			t.history = t.history[len(t.history)-1000:]
		}
	}
}

// GetStats returns current statistics.
func (t *TokenTracker) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	savingsRate := 0.0
	if t.totalInput > 0 {
		savingsRate = float64(t.totalSaved) / float64(t.totalInput) * 100
	}

	return map[string]interface{}{
		"total_saved":   t.totalSaved,
		"total_input":   t.totalInput,
		"total_output":  t.totalOutput,
		"savings_rate":  savingsRate,
		"session_start": t.sessionStart.Unix(),
		"uptime":        time.Since(t.sessionStart).Seconds(),
	}
}

// GetHistory returns historical data.
func (t *TokenTracker) GetHistory() []DataPoint {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]DataPoint, len(t.history))
	copy(result, t.history)
	return result
}

// NewLiveTicker creates a new live ticker.
func NewLiveTicker(server *LiveServer, tracker *TokenTracker) *LiveTicker {
	return &LiveTicker{
		server:   server,
		tracker:  tracker,
		stop:     make(chan bool),
		interval: 5 * time.Second,
	}
}

// Start starts the ticker.
func (t *LiveTicker) Start() {
	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.broadcast()
			case <-t.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the ticker.
func (t *LiveTicker) Stop() {
	close(t.stop)
}

func (t *LiveTicker) broadcast() {
	stats := t.tracker.GetStats()
	t.server.Broadcast("ticker", stats)
}

// HeatmapData represents command frequency heatmap data.
type HeatmapData struct {
	Cells []HeatmapCell `json:"cells"`
	Max   int           `json:"max"`
}

// HeatmapCell represents a single cell in the heatmap.
type HeatmapCell struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Value int    `json:"value"`
	Label string `json:"label,omitempty"`
}

// GenerateHeatmap generates heatmap data from command history.
func GenerateHeatmap(commands []string, window time.Duration) HeatmapData {
	// Group by day of week and hour of day
	days := make([][]int, 7)
	for i := range days {
		days[i] = make([]int, 24)
	}

	// Parse commands and count
	// This is a simplified implementation
	// In production, parse timestamps from commands

	max := 0
	var cells []HeatmapCell

	for day := 0; day < 7; day++ {
		for hour := 0; hour < 24; hour++ {
			value := days[day][hour]
			if value > 0 {
				cells = append(cells, HeatmapCell{
					X:     hour,
					Y:     day,
					Value: value,
				})
				if value > max {
					max = value
				}
			}
		}
	}

	return HeatmapData{
		Cells: cells,
		Max:   max,
	}
}

// generateClientID generates a unique client ID.
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
