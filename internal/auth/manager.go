package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenNotFound      = errors.New("token not found")
)

// AuthManager handles user authentication and authorization.
type AuthManager struct {
	db *sql.DB
}

// NewAuthManager creates a new auth manager.
func NewAuthManager(db *sql.DB) *AuthManager {
	return &AuthManager{db: db}
}

// ValidateToken validates an auth token and returns the auth context.
func (am *AuthManager) ValidateToken(ctx context.Context, token string) (*AuthContext, error) {
	var (
		userID    string
		teamID    string
		role      UserRole
		email     string
		name      string
		isActive  bool
		expiresAt time.Time
	)

	row := am.db.QueryRowContext(ctx, `
		SELECT u.id, u.team_id, u.role, u.email, u.full_name, u.is_active, s.expires_at
		FROM user_sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = ? AND s.expires_at > NOW()
	`, token)

	err := row.Scan(&userID, &teamID, &role, &email, &name, &isActive, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}

	// Update last used time
	_, _ = am.db.ExecContext(ctx, `
		UPDATE user_sessions SET last_used_at = NOW() WHERE token = ?
	`, token)

	return &AuthContext{
		UserID:    userID,
		TeamID:    teamID,
		Role:      role,
		Email:     email,
		Name:      name,
		IsActive:  isActive,
		ExpiresAt: expiresAt,
	}, nil
}

// CreateSession creates a new user session.
func (am *AuthManager) CreateSession(ctx context.Context, userID, teamID string, duration time.Duration) (*UserSession, error) {
	token := generateToken()
	expiresAt := time.Now().Add(duration)

	result, err := am.db.ExecContext(ctx, `
		INSERT INTO user_sessions (user_id, team_id, token, expires_at, created_at, last_used_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`, userID, teamID, token, expiresAt)

	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	return &UserSession{
		ID:        string(rune(id)),
		UserID:    userID,
		TeamID:    teamID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		LastUsedAt: time.Now(),
	}, nil
}

// RevokeSession revokes a user session.
func (am *AuthManager) RevokeSession(ctx context.Context, token string) error {
	_, err := am.db.ExecContext(ctx, `
		DELETE FROM user_sessions WHERE token = ?
	`, token)
	return err
}

// RevokeAllUserSessions revokes all sessions for a user.
func (am *AuthManager) RevokeAllUserSessions(ctx context.Context, userID string) error {
	_, err := am.db.ExecContext(ctx, `
		DELETE FROM user_sessions WHERE user_id = ?
	`, userID)
	return err
}

// EnforcePermission returns an error if the auth context doesn't have the required permission.
func EnforcePermission(auth *AuthContext, perm Permission) error {
	if auth == nil || !auth.IsValid() {
		return ErrUnauthorized
	}
	if !auth.HasPermission(perm) {
		return ErrForbidden
	}
	return nil
}

// EnforceTeam returns an error if the auth context is not for the specified team.
func EnforceTeam(auth *AuthContext, teamID string) error {
	if auth == nil || !auth.IsValid() {
		return ErrUnauthorized
	}
	if auth.TeamID != teamID {
		return ErrForbidden
	}
	return nil
}

// EnforceTeamAndPermission returns an error if the auth context is not for the specified team
// and doesn't have the required permission.
func EnforceTeamAndPermission(auth *AuthContext, teamID string, perm Permission) error {
	if err := EnforceTeam(auth, teamID); err != nil {
		return err
	}
	if err := EnforcePermission(auth, perm); err != nil {
		return err
	}
	return nil
}

// generateToken generates a random token.
func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
