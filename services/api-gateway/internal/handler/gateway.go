package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	compressionpb "github.com/GrayCodeAI/tokman/services/compression-service/proto/compression"
	securitypb "github.com/GrayCodeAI/tokman/services/security-service/proto/security"
	commonpb "github.com/GrayCodeAI/tokman/services/shared/proto/common"
)

// Gateway handles HTTP requests and routes to gRPC services
type Gateway struct {
	compressionClient compressionpb.CompressionServiceClient
	securityClient    securitypb.SecurityServiceClient
	connCompression   *grpc.ClientConn
	connSecurity      *grpc.ClientConn
}

// NewGateway creates a new API gateway
func NewGateway(compressionAddr, analyticsAddr, securityAddr, configAddr string) (*Gateway, error) {
	// Connect to compression service
	connCompression, err := grpc.Dial(compressionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to compression service: %w", err)
	}

	// Connect to security service
	connSecurity, err := grpc.Dial(securityAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		connCompression.Close()
		return nil, fmt.Errorf("failed to connect to security service: %w", err)
	}

	return &Gateway{
		compressionClient: compressionpb.NewCompressionServiceClient(connCompression),
		securityClient:    securitypb.NewSecurityServiceClient(connSecurity),
		connCompression:   connCompression,
		connSecurity:      connSecurity,
	}, nil
}

// Close closes all service connections
func (g *Gateway) Close() error {
	if g.connCompression != nil {
		g.connCompression.Close()
	}
	if g.connSecurity != nil {
		g.connSecurity.Close()
	}
	return nil
}

// CompressRequest represents a compression request
type CompressRequest struct {
	Text         string            `json:"text"`
	Mode         string            `json:"mode,omitempty"`
	Budget       int32             `json:"budget,omitempty"`
	QueryIntent  string            `json:"query,omitempty"`
	Preset       string            `json:"preset,omitempty"`
	EnableLayers map[string]bool   `json:"enable_layers,omitempty"`
	DisableLayers map[string]bool  `json:"disable_layers,omitempty"`
}

// CompressHandler handles compression requests
func (g *Gateway) CompressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
			return
		}

		var req CompressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "MISSING_TEXT", "text field is required")
			return
		}

		// Security scan first
		scanResp, err := g.securityClient.ScanContent(r.Context(), &securitypb.ScanRequest{
			Content: req.Text,
		})
		if err != nil {
			slog.Error("Security scan failed", "error", err)
			// Continue anyway, just log
		} else if scanResp.HasCritical {
			writeError(w, http.StatusBadRequest, "SECURITY_VIOLATION", "Content contains critical security findings")
			return
		}

		// Convert mode string to enum
		mode := commonpb.CompressionMode_MODE_MINIMAL
		switch req.Mode {
		case "aggressive":
			mode = commonpb.CompressionMode_MODE_AGGRESSIVE
		case "none":
			mode = commonpb.CompressionMode_MODE_NONE
		}

		// Convert preset string to enum
		var preset commonpb.PipelinePreset
		switch req.Preset {
		case "fast":
			preset = commonpb.PipelinePreset_PRESET_FAST
		case "balanced":
			preset = commonpb.PipelinePreset_PRESET_BALANCED
		case "full":
			preset = commonpb.PipelinePreset_PRESET_FULL
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		resp, err := g.compressionClient.Compress(ctx, &compressionpb.CompressRequest{
			Text:          req.Text,
			Mode:          mode,
			Budget:        req.Budget,
			QueryIntent:   req.QueryIntent,
			Preset:        preset,
			EnableLayers:  req.EnableLayers,
			DisableLayers: req.DisableLayers,
		})
		if err != nil {
			slog.Error("Compression failed", "error", err)
			writeError(w, http.StatusInternalServerError, "COMPRESSION_FAILED", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// CompressStreamHandler handles streaming compression
func (g *Gateway) CompressStreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Streaming compression not yet implemented")
	}
}

// ScanHandler handles security scan requests
func (g *Gateway) ScanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "READ_ERROR", err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := g.securityClient.ScanContent(ctx, &securitypb.ScanRequest{
			Content: string(body),
		})
		if err != nil {
			slog.Error("Security scan failed", "error", err)
			writeError(w, http.StatusInternalServerError, "SCAN_FAILED", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// RedactHandler handles PII redaction requests
func (g *Gateway) RedactHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "READ_ERROR", err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := g.securityClient.RedactPII(ctx, &securitypb.RedactRequest{
			Content: string(body),
		})
		if err != nil {
			slog.Error("Redaction failed", "error", err)
			writeError(w, http.StatusInternalServerError, "REDACT_FAILED", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// StatsHandler handles stats requests
func (g *Gateway) StatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := g.compressionClient.GetStats(ctx, &compressionpb.StatsRequest{})
		if err != nil {
			slog.Error("GetStats failed", "error", err)
			writeError(w, http.StatusInternalServerError, "STATS_FAILED", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// CommandsHandler handles command history requests
func (g *Gateway) CommandsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Command history not yet implemented")
	}
}

// SavingsHandler handles savings requests
func (g *Gateway) SavingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Savings tracking not yet implemented")
	}
}

// FiltersHandler handles filter listing requests
func (g *Gateway) FiltersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		resp, err := g.compressionClient.ListFilters(ctx, &commonpb.Empty{})
		if err != nil {
			slog.Error("ListFilters failed", "error", err)
			writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// HealthHandler handles health checks
func (g *Gateway) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":    "healthy",
			"service":   "api-gateway",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// ServicesHealthHandler checks health of all backend services
func (g *Gateway) ServicesHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		health := map[string]interface{}{
			"api-gateway": map[string]string{
				"status": "healthy",
			},
		}

		// Check compression service
		if resp, err := g.compressionClient.Health(ctx, &commonpb.HealthRequest{Service: "compression"}); err == nil {
			health["compression-service"] = map[string]interface{}{
				"status": resp.Status.String(),
			}
		} else {
			health["compression-service"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		}

		// Check security service
		if resp, err := g.securityClient.Health(ctx, &commonpb.HealthRequest{Service: "security"}); err == nil {
			health["security-service"] = map[string]interface{}{
				"status": resp.Status.String(),
			}
		} else {
			health["security-service"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		}

		writeJSON(w, http.StatusOK, health)
	}
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	writeJSON(w, code, map[string]interface{}{
		"error":   errCode,
		"message": message,
		"status":  code,
	})
}
