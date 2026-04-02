package mcphost

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHost(t *testing.T) {
	config := HostConfig{
		Name:    "TestHost",
		Version: "1.0",
	}
	host := NewHost(config)

	if host == nil {
		t.Fatal("expected host to be created")
	}

	if host.name != "TestHost" {
		t.Errorf("expected name 'TestHost', got %s", host.name)
	}

	if host.servers == nil {
		t.Error("expected servers map to be initialized")
	}

	if host.sessions == nil {
		t.Error("expected sessions map to be initialized")
	}
}

func TestHostID(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	if host.ID() == "" {
		t.Error("expected host ID to be generated")
	}
}

func TestHostName(t *testing.T) {
	config := HostConfig{Name: "TestHost"}
	host := NewHost(config)

	if host.Name() != "TestHost" {
		t.Errorf("expected 'TestHost', got %s", host.Name())
	}
}

func TestRegisterServer(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	server, err := host.RegisterServer("server-1", "TestServer", nil)
	if err != nil {
		t.Fatalf("failed to register server: %v", err)
	}

	if server == nil {
		t.Fatal("expected server to be created")
	}

	if server.ID != "server-1" {
		t.Errorf("expected ID 'server-1', got %s", server.ID)
	}

	if server.Name != "TestServer" {
		t.Errorf("expected name 'TestServer', got %s", server.Name)
	}

	// Register duplicate
	_, err = host.RegisterServer("server-1", "TestServer", nil)
	if err == nil {
		t.Error("expected error for duplicate server")
	}
}

func TestListServers(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	host.RegisterServer("s1", "Server1", nil)
	host.RegisterServer("s2", "Server2", nil)

	servers := host.ListServers()

	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestGetServer(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	host.RegisterServer("s1", "Server1", nil)

	server, err := host.GetServer("s1")
	if err != nil {
		t.Fatalf("failed to get server: %v", err)
	}

	if server.ID != "s1" {
		t.Errorf("expected ID 's1', got %s", server.ID)
	}

	_, err = host.GetServer("missing")
	if err == nil {
		t.Error("expected error for missing server")
	}
}

func TestCreateSession(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	host.RegisterServer("s1", "Server1", nil)

	clientInfo := ClientInfo{Name: "TestClient", Version: "1.0"}
	session, err := host.CreateSession(clientInfo, "s1")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session == nil {
		t.Fatal("expected session to be created")
	}

	if session.ClientInfo.Name != "TestClient" {
		t.Errorf("expected client name 'TestClient', got %s", session.ClientInfo.Name)
	}

	if session.Status != SessionStatusActive {
		t.Errorf("expected status 'active', got %s", session.Status)
	}

	// Create session for non-existent server
	_, err = host.CreateSession(clientInfo, "missing")
	if err == nil {
		t.Error("expected error for missing server")
	}
}

func TestCloseSession(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	host.RegisterServer("s1", "Server1", nil)

	clientInfo := ClientInfo{Name: "TestClient", Version: "1.0"}
	session, _ := host.CreateSession(clientInfo, "s1")

	err := host.CloseSession(session.ID)
	if err != nil {
		t.Fatalf("failed to close session: %v", err)
	}

	// Try to close non-existent session
	err = host.CloseSession("missing")
	if err == nil {
		t.Error("expected error for missing session")
	}
}

func TestShutdown(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	host.RegisterServer("s1", "Server1", nil)
	host.RegisterServer("s2", "Server2", nil)

	err := host.Shutdown()
	if err != nil {
		t.Fatalf("failed to shutdown: %v", err)
	}
}

func TestHostEventHandler(t *testing.T) {
	config := HostConfig{Name: "Test"}
	host := NewHost(config)

	events := make([]HostEvent, 0)
	handler := func(e HostEvent) {
		events = append(events, e)
	}

	host.SetEventHandler(handler)

	host.RegisterServer("s1", "Server1", nil)

	clientInfo := ClientInfo{Name: "TestClient", Version: "1.0"}
	host.CreateSession(clientInfo, "s1")

	time.Sleep(10 * time.Millisecond)

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestMessageJSON(t *testing.T) {
	msg := Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"test": true}`),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Errorf("expected JSONRPC '2.0', got %s", decoded.JSONRPC)
	}
}

func TestErrorJSON(t *testing.T) {
	err := &Error{
		Code:    -32600,
		Message: "Invalid Request",
	}

	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("failed to marshal: %v", marshalErr)
	}

	var decoded Error
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Code != -32600 {
		t.Errorf("expected code -32600, got %d", decoded.Code)
	}
}

func BenchmarkRegisterServer(b *testing.B) {
	config := HostConfig{Name: "Bench"}
	host := NewHost(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		host.RegisterServer("s", "Server", nil)
	}
}
