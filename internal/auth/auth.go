package auth

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Config holds authentication configuration.
type Config struct {
	Enabled bool
	// Future: TokenValidator, APIKeyStore, etc.
}

// Interceptor returns a gRPC unary interceptor for authentication.
// When auth is disabled (default for local use), all requests pass through.
func Interceptor(cfg Config, logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !cfg.Enabled {
			// Auth disabled - allow all requests (local-only default)
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("auth: missing metadata", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			logger.Warn("auth: missing authorization header", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// TODO: Implement actual token validation
		// For now, just log and reject all auth attempts when enabled
		logger.Warn("auth: token validation not yet implemented",
			"method", info.FullMethod,
			"token_prefix", truncateToken(tokens[0]),
		)
		return nil, status.Error(codes.Unimplemented, "auth not yet implemented - disable auth for local use")
	}
}

// StreamInterceptor returns a gRPC stream interceptor for authentication.
func StreamInterceptor(cfg Config, logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !cfg.Enabled {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			logger.Warn("auth: missing metadata", "method", info.FullMethod)
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			logger.Warn("auth: missing authorization header", "method", info.FullMethod)
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		logger.Warn("auth: token validation not yet implemented",
			"method", info.FullMethod,
			"token_prefix", truncateToken(tokens[0]),
		)
		return status.Error(codes.Unimplemented, "auth not yet implemented - disable auth for local use")
	}
}

func truncateToken(token string) string {
	if len(token) > 10 {
		return token[:10] + "..."
	}
	return token
}
