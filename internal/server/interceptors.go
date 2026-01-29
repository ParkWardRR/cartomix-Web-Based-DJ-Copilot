package server

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor logs all unary RPC calls with timing and status.
func UnaryLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the call
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		logger.Info("gRPC unary call",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"code", code.String(),
		)

		return resp, err
	}
}

// StreamLoggingInterceptor logs all streaming RPC calls.
func StreamLoggingInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Log the call
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		logger.Info("gRPC stream call",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"code", code.String(),
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		return err
	}
}

// RecoveryInterceptor recovers from panics and returns an Internal error.
func RecoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC handler panic recovered",
					"method", info.FullMethod,
					"panic", r,
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor recovers from panics in streaming handlers.
func StreamRecoveryInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC stream handler panic recovered",
					"method", info.FullMethod,
					"panic", r,
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}

// Metrics tracks basic gRPC metrics.
type Metrics struct {
	TotalRequests     int64
	TotalErrors       int64
	StreamsOpened     int64
	StreamsClosed     int64
	TotalLatencyMs    int64
	RequestsByMethod  map[string]int64
	ErrorsByMethod    map[string]int64
	LatencyByMethod   map[string]int64
}

var globalMetrics = &Metrics{
	RequestsByMethod: make(map[string]int64),
	ErrorsByMethod:   make(map[string]int64),
	LatencyByMethod:  make(map[string]int64),
}

// GetMetrics returns the current metrics snapshot.
func GetMetrics() Metrics {
	return *globalMetrics
}

// MetricsInterceptor collects basic metrics for unary calls.
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Milliseconds()

		globalMetrics.TotalRequests++
		globalMetrics.TotalLatencyMs += duration
		globalMetrics.RequestsByMethod[info.FullMethod]++
		globalMetrics.LatencyByMethod[info.FullMethod] += duration

		if err != nil {
			globalMetrics.TotalErrors++
			globalMetrics.ErrorsByMethod[info.FullMethod]++
		}

		return resp, err
	}
}

// StreamMetricsInterceptor collects basic metrics for streaming calls.
func StreamMetricsInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		globalMetrics.StreamsOpened++

		err := handler(srv, ss)

		globalMetrics.StreamsClosed++
		duration := time.Since(start).Milliseconds()

		globalMetrics.TotalRequests++
		globalMetrics.TotalLatencyMs += duration
		globalMetrics.RequestsByMethod[info.FullMethod]++
		globalMetrics.LatencyByMethod[info.FullMethod] += duration

		if err != nil {
			globalMetrics.TotalErrors++
			globalMetrics.ErrorsByMethod[info.FullMethod]++
		}

		return err
	}
}

// ChainUnaryInterceptors chains multiple unary interceptors.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(ctx context.Context, req interface{}) (interface{}, error) {
				return interceptor(ctx, req, info, next)
			}
		}
		return chain(ctx, req)
	}
}

// ChainStreamInterceptors chains multiple stream interceptors.
func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(srv interface{}, ss grpc.ServerStream) error {
				return interceptor(srv, ss, info, next)
			}
		}
		return chain(srv, ss)
	}
}
