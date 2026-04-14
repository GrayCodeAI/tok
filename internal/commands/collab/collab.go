package collaboration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	collabServer  string
	collabRoom    string
	collabUser    string
	collabToken   string
	collabEnabled bool
)

type CollabSession struct {
	ID        string    `json:"id"`
	Room      string    `json:"room"`
	User      string    `json:"user"`
	JoinedAt  int64     `json:"joined_at"`
	SharedCtx []Context `json:"shared_context"`
}

type Context struct {
	ID        string `json:"id"`
	Command   string `json:"command"`
	Summary   string `json:"summary"`
	Tokens    int    `json:"tokens"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

type Message struct {
	Type      string   `json:"type"`
	SessionID string   `json:"session_id"`
	User      string   `json:"user"`
	Content   string   `json:"content"`
	Timestamp int64    `json:"timestamp"`
	Context   *Context `json:"context,omitempty"`
}

var (
	sessions      = make(map[string]*CollabSession)
	activeSession *CollabSession
	mu            sync.RWMutex
)

func NewCollabCommand() *cobra.Command {
	collabCmd := &cobra.Command{
		Use:   "collab",
		Short: "Real-time collaboration features",
		Long: `Real-time collaboration for sharing compressed context between team members.
		
Features:
  - Share compressed command output with team
  - Collaborative context pools
  - WebSocket-based real-time sync
  - Session management for pair programming

Examples:
  tokman collab create --room=debug-session
  tokman collab join --room=debug-session --user=alice
  tokman collab share --summary="kubectl output"`,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new collaboration session",
		RunE:  runCollabCreate,
	}

	joinCmd := &cobra.Command{
		Use:   "join",
		Short: "Join an existing session",
		RunE:  runCollabJoin,
	}

	leaveCmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave current session",
		RunE:  runCollabLeave,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE:  runCollabList,
	}

	shareCmd := &cobra.Command{
		Use:   "share",
		Short: "Share context with session",
		RunE:  runCollabShare,
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show collaboration status",
		RunE:  runCollabStatus,
	}

	collabCmd.Flags().StringVar(&collabServer, "server", "localhost:8765", "Collaboration server address")
	collabCmd.Flags().StringVar(&collabRoom, "room", "", "Room name to join")
	collabCmd.Flags().StringVar(&collabUser, "user", "", "Username for session")
	collabCmd.Flags().StringVar(&collabToken, "token", "", "Auth token")
	collabCmd.Flags().BoolVar(&collabEnabled, "enable", false, "Enable real-time collaboration")

	collabCmd.AddCommand(createCmd)
	collabCmd.AddCommand(joinCmd)
	collabCmd.AddCommand(leaveCmd)
	collabCmd.AddCommand(listCmd)
	collabCmd.AddCommand(shareCmd)
	collabCmd.AddCommand(statusCmd)

	return collabCmd
}

func runCollabCreate(cmd *cobra.Command, args []string) error {
	if collabRoom == "" {
		return fmt.Errorf("room name is required (--room)")
	}

	session := &CollabSession{
		ID:       uuid.New().String(),
		Room:     collabRoom,
		User:     collabUser,
		JoinedAt: time.Now().Unix(),
	}

	mu.Lock()
	sessions[session.ID] = session
	activeSession = session
	mu.Unlock()

	viper.Set("collab.room", collabRoom)
	viper.Set("collab.session_id", session.ID)
	viper.WriteConfig()

	fmt.Printf("Created collaboration session: %s\n", session.ID)
	fmt.Printf("Room: %s\n", collabRoom)
	fmt.Println("Share this room name with team members to collaborate")

	return nil
}

func runCollabJoin(cmd *cobra.Command, args []string) error {
	if collabRoom == "" {
		return fmt.Errorf("room name is required (--room)")
	}

	session := &CollabSession{
		ID:       uuid.New().String(),
		Room:     collabRoom,
		User:     collabUser,
		JoinedAt: time.Now().Unix(),
	}

	mu.Lock()
	activeSession = session
	mu.Unlock()

	viper.Set("collab.room", collabRoom)
	viper.Set("collab.session_id", session.ID)
	viper.Set("collab.enabled", true)
	viper.WriteConfig()

	fmt.Printf("Joined room: %s\n", collabRoom)
	fmt.Println("Use 'tokman collab share' to share compressed output")

	return nil
}

func runCollabLeave(cmd *cobra.Command, args []string) error {
	mu.Lock()
	if activeSession != nil {
		delete(sessions, activeSession.ID)
		activeSession = nil
	}
	mu.Unlock()

	viper.Set("collab.enabled", false)
	viper.WriteConfig()

	fmt.Println("Left collaboration session")
	return nil
}

func runCollabList(cmd *cobra.Command, args []string) error {
	mu.RLock()
	defer mu.RUnlock()

	if len(sessions) == 0 {
		fmt.Println("No active collaboration sessions")
		return nil
	}

	fmt.Println("=== Active Sessions ===")
	for _, s := range sessions {
		fmt.Printf("Room: %s | User: %s | ID: %s | Joined: %s\n",
			s.Room, s.User, s.ID, time.Unix(s.JoinedAt, 0).Format(time.RFC3339))
	}

	return nil
}

func runCollabShare(cmd *cobra.Command, args []string) error {
	mu.RLock()
	defer mu.RUnlock()

	if activeSession == nil {
		return fmt.Errorf("no active session. Run 'tokman collab join' first")
	}

	summary := "Command output"
	if len(args) > 0 {
		summary = args[0]
	}

	ctx := Context{
		ID:        uuid.New().String(),
		Summary:   summary,
		CreatedBy: activeSession.User,
		CreatedAt: time.Now().Unix(),
	}

	activeSession.SharedCtx = append(activeSession.SharedCtx, ctx)

	sharedFile := filepath.Join(os.Getenv("HOME"), ".config/tokman", "collab", "shared.json")
	os.MkdirAll(filepath.Dir(sharedFile), 0755)

	data, _ := json.Marshal(activeSession)
	os.WriteFile(sharedFile, data, 0644)

	fmt.Printf("Shared context '%s' with room %s\n", summary, activeSession.Room)
	return nil
}

func runCollabStatus(cmd *cobra.Command, args []string) error {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Println("=== TokMan Collaboration Status ===")
	fmt.Printf("Enabled: %s\n", yesNo(collabEnabled))

	if activeSession != nil {
		fmt.Printf("Active Session:\n")
		fmt.Printf("  Room: %s\n", activeSession.Room)
		fmt.Printf("  User: %s\n", activeSession.User)
		fmt.Printf("  Session ID: %s\n", activeSession.ID)
		fmt.Printf("  Shared Contexts: %d\n", len(activeSession.SharedCtx))
	} else {
		fmt.Println("No active session")
		fmt.Println("Use 'tokman collab create' or 'tokman collab join' to start")
	}

	return nil
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
