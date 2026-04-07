package auth

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCInterceptor provides gRPC interceptors for authentication and authorization.
type GRPCInterceptor struct {
	authManager *AuthManager
	logger      *slog.Logger
}

// NewGRPCInterceptor creates a new gRPC interceptor.
func NewGRPCInterceptor(authManager *AuthManager, logger *slog.Logger) *GRPCInterceptor {
	if logger == nil {
		logger = slog.Default()
	}
	return &GRPCInterceptor{
		authManager: authManager,
		logger:      logger,
	}
}

// UnaryInterceptor returns a unary server interceptor for authentication.
func (gi *GRPCInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract auth token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		auth := md.Get("authorization")
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Extract token from "Bearer <token>" format
		token := extractToken(auth[0])
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Validate token
		authCtx, err := gi.authManager.ValidateToken(ctx, token)
		if err != nil {
			gi.logger.Debug("failed to validate token", "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Add auth context to request context
		newCtx := context.WithValue(ctx, ContextKeyAuth, authCtx)

		gi.logger.Debug("request authorized", "user_id", authCtx.UserID, "team_id", authCtx.TeamID, "method", info.FullMethod)

		return handler(newCtx, req)
	}
}

// StreamInterceptor returns a stream server interceptor for authentication.
func (gi *GRPCInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Extract auth token from metadata
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		auth := md.Get("authorization")
		if len(auth) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Extract token from "Bearer <token>" format
		token := extractToken(auth[0])
		if token == "" {
			return status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Validate token
		authCtx, err := gi.authManager.ValidateToken(ss.Context(), token)
		if err != nil {
			gi.logger.Debug("failed to validate token", "error", err)
			return status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Create a wrapper stream with auth context
		wrappedStream := &authStreamWrapper{
			ServerStream: ss,
			ctx:          context.WithValue(ss.Context(), ContextKeyAuth, authCtx),
		}

		gi.logger.Debug("stream authorized", "user_id", authCtx.UserID, "team_id", authCtx.TeamID, "method", info.FullMethod)

		return handler(srv, wrappedStream)
	}
}

// authStreamWrapper wraps grpc.ServerStream to provide custom context.
type authStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *authStreamWrapper) Context() context.Context {
	return w.ctx
}

// ContextKey is used for storing auth context in request context.
type ContextKey string

const ContextKeyAuth ContextKey = "auth_context"

// extractToken extracts the token from an authorization header.
// Expected format: "Bearer <token>"
func extractToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

// GetAuthContext retrieves the auth context from a request context.
func GetAuthContext(ctx context.Context) *AuthContext {
	authCtx, ok := ctx.Value(ContextKeyAuth).(*AuthContext)
	if !ok {
		return nil
	}
	return authCtx
}

// MustGetAuthContext retrieves the auth context or panics if not found.
func MustGetAuthContext(ctx context.Context) *AuthContext {
	authCtx := GetAuthContext(ctx)
	if authCtx == nil {
		panic("auth context not found in request context")
	}
	return authCtx
}
