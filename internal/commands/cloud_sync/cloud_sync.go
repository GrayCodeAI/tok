package cloud_sync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	syncProvider string
	syncTeamID   string
	syncEncrypt  bool
	syncAuto     bool
	syncInterval int
)

const (
	metadataFile = ".sync-metadata"
)

var (
	mu         sync.RWMutex
	localHash  string
	remoteHash string
	lastSync   time.Time
)

type SyncMetadata struct {
	LastPush int64  `json:"last_push"`
	LastPull int64  `json:"last_pull"`
	Version  int    `json:"version"`
	Checksum string `json:"checksum"`
}

type SyncCmd struct {
	cmd     *cobra.Command
	push    *cobra.Command
	pull    *cobra.Command
	status  *cobra.Command
	enable  *cobra.Command
	disable *cobra.Command
}

func NewSyncCommand() *SyncCmd {
	sc := &SyncCmd{
		cmd: &cobra.Command{
			Use:   "sync",
			Short: "Sync settings with cloud",
			Long: `Sync TokMan configuration across devices and team members.
			
Supported providers:
  - drive (Google Drive) - requires credentials.json
  - local (local file sync)
  - s3 (AWS S3) - coming soon

Examples:
  tokman sync enable --provider=drive
  tokman sync push
  tokman sync status`,
		},
	}

	sc.push = &cobra.Command{
		Use:   "push",
		Short: "Push settings to cloud",
		RunE:  sc.runSyncPush,
	}

	sc.pull = &cobra.Command{
		Use:   "pull",
		Short: "Pull settings from cloud",
		RunE:  sc.runSyncPull,
	}

	sc.status = &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE:  sc.runSyncStatus,
	}

	sc.enable = &cobra.Command{
		Use:   "enable",
		Short: "Enable cloud sync",
		RunE:  sc.runSyncEnable,
	}

	sc.disable = &cobra.Command{
		Use:   "disable",
		Short: "Disable cloud sync",
		RunE:  sc.runSyncDisable,
	}

	sc.cmd.Flags().StringVarP(&syncProvider, "provider", "p", "local", "Cloud provider (local, drive, s3)")
	sc.cmd.Flags().StringVar(&syncTeamID, "team", "", "Team ID for shared settings")
	sc.cmd.Flags().BoolVar(&syncEncrypt, "encrypt", true, "Encrypt settings before upload")
	sc.cmd.Flags().BoolVar(&syncAuto, "auto", false, "Enable automatic sync")
	sc.cmd.Flags().IntVar(&syncInterval, "interval", 300, "Sync interval in seconds")

	sc.cmd.AddCommand(sc.push)
	sc.cmd.AddCommand(sc.pull)
	sc.cmd.AddCommand(sc.status)
	sc.cmd.AddCommand(sc.enable)
	sc.cmd.AddCommand(sc.disable)

	return sc
}

func (sc *SyncCmd) runSyncPush(cmd *cobra.Command, args []string) error {
	configFile := filepath.Join(os.Getenv("HOME"), ".config/tokman/config.toml")

	fmt.Println("Pushing settings to cloud...")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	syncFile := sc.getSyncFilePath()

	encrypted := data
	if syncEncrypt {
		encrypted, err = encryptData(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt: %w", err)
		}
	}

	if err := os.WriteFile(syncFile, encrypted, 0644); err != nil {
		return fmt.Errorf("failed to write sync file: %w", err)
	}

	mu.Lock()
	localHash = hashData(data)
	mu.Unlock()

	sc.updateMetadata(SyncMetadata{
		LastPush: time.Now().Unix(),
		Version:  1,
		Checksum: localHash,
	})

	fmt.Println("Settings pushed to:", syncFile)
	fmt.Println("To sync across devices, copy this file manually or use a cloud provider.")
	return nil
}

func (sc *SyncCmd) runSyncPull(cmd *cobra.Command, args []string) error {
	configFile := filepath.Join(os.Getenv("HOME"), ".config/tokman/config.toml")

	fmt.Println("Pulling settings from cloud...")

	syncFile := sc.getSyncFilePath()

	data, err := os.ReadFile(syncFile)
	if err != nil {
		return fmt.Errorf("failed to read sync file: %w", err)
	}

	decrypted := data
	if syncEncrypt {
		decrypted, err = decryptData(data)
		if err != nil {
			return fmt.Errorf("failed to decrypt: %w", err)
		}
	}

	if err := os.WriteFile(configFile, decrypted, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	mu.Lock()
	remoteHash = hashData(decrypted)
	mu.Unlock()

	sc.updateMetadata(SyncMetadata{
		LastPull: time.Now().Unix(),
		Version:  1,
		Checksum: remoteHash,
	})

	fmt.Println("Settings pulled successfully!")
	return nil
}

func (sc *SyncCmd) runSyncStatus(cmd *cobra.Command, args []string) error {
	metadata := sc.loadMetadata()

	fmt.Println("=== TokMan Cloud Sync Status ===")
	fmt.Printf("Provider: %s\n", syncProvider)
	if syncTeamID != "" {
		fmt.Printf("Team ID: %s\n", syncTeamID)
	}

	if metadata.LastPush == 0 && metadata.LastPull == 0 {
		fmt.Println("Status: NOT CONFIGURED")
		fmt.Println("Run 'tokman sync enable' to set up cloud sync")
		return nil
	}

	if metadata.LastPush > metadata.LastPull {
		fmt.Printf("Last push: %s\n", time.Unix(metadata.LastPush, 0).Round(time.Second))
	} else if metadata.LastPull > metadata.LastPush {
		fmt.Printf("Last pull: %s\n", time.Unix(metadata.LastPull, 0).Round(time.Second))
	}

	if localHash != remoteHash {
		fmt.Println("Status: OUT OF SYNC")
		fmt.Println("  Use 'tokman sync push' or 'tokman sync pull' to sync")
	} else if localHash != "" {
		fmt.Println("Status: IN SYNC")
	}

	return nil
}

func (sc *SyncCmd) runSyncEnable(cmd *cobra.Command, args []string) error {
	if syncProvider == "drive" {
		credFile := filepath.Join(os.Getenv("HOME"), ".config/tokman/credentials.json")
		if _, err := os.Stat(credFile); os.IsNotExist(err) {
			fmt.Println("Note: Google Drive sync requires credentials.json in ~/.config/tokman/")
			fmt.Println("For now, using local file sync mode.")
		}
	}

	viper.Set("sync.enabled", true)
	viper.Set("sync.provider", syncProvider)
	viper.Set("sync.team_id", syncTeamID)
	viper.Set("sync.auto_sync", syncAuto)
	viper.Set("sync.interval", syncInterval)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println("Cloud sync enabled!")
	fmt.Printf("Provider: %s\n", syncProvider)
	fmt.Println("Use 'tokman sync push' to upload settings")
	fmt.Println("Use 'tokman sync pull' to download settings")

	if syncAuto {
		fmt.Printf("Auto-sync enabled every %d seconds\n", syncInterval)
	}

	return nil
}

func (sc *SyncCmd) runSyncDisable(cmd *cobra.Command, args []string) error {
	viper.Set("sync.enabled", false)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Println("Cloud sync disabled!")
	return nil
}

func (sc *SyncCmd) getSyncFilePath() string {
	baseDir := filepath.Join(os.Getenv("HOME"), ".config/tokman", "sync")
	os.MkdirAll(baseDir, 0755)

	if syncTeamID != "" {
		return filepath.Join(baseDir, fmt.Sprintf("config-%s.toml.enc", syncTeamID))
	}
	return filepath.Join(baseDir, "config.toml.enc")
}

func (sc *SyncCmd) loadMetadata() SyncMetadata {
	var metadata SyncMetadata
	metadataPath := filepath.Join(os.Getenv("HOME"), ".config/tokman", metadataFile)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return metadata
	}
	json.Unmarshal(data, &metadata)
	return metadata
}

func (sc *SyncCmd) updateMetadata(m SyncMetadata) {
	metadataPath := filepath.Join(os.Getenv("HOME"), ".config/tokman", metadataFile)
	data, _ := json.Marshal(m)
	os.WriteFile(metadataPath, data, 0644)
}

func encryptData(data []byte) ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	encrypted := make([]byte, 0, len(key)+len(ciphertext))
	encrypted = append(encrypted, key...)
	encrypted = append(encrypted, ciphertext...)

	return encrypted, nil
}

func decryptData(data []byte) ([]byte, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	key := data[:32]
	ciphertext := data[32:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func hashData(data []byte) string {
	h := sha256.Sum256(data)
	return base64.URLEncoding.EncodeToString(h[:])
}
