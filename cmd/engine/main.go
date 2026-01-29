package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/cartomix/cancun/gen/go/engine"
	"github.com/cartomix/cancun/internal/analyzer"
	"github.com/cartomix/cancun/internal/auth"
	"github.com/cartomix/cancun/internal/config"
	"github.com/cartomix/cancun/internal/server"
	"github.com/cartomix/cancun/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Parse()

	// Setup structured logger
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		logger.Error("failed to create data directory", "path", cfg.DataDir, "error", err)
		os.Exit(1)
	}

	// Open database
	db, err := storage.Open(cfg.DataDir, logger)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Prefer Swift analyzer worker; fall back to CPU placeholder.
	var analysisBackend analyzer.Analyzer
	analysisBackend, err = analyzer.NewClient(cfg.AnalyzerAddr, logger)
	if err != nil {
		logger.Warn("analyzer worker unavailable, falling back to CPU", "error", err)
		analysisBackend = analyzer.NewCPUFallback(logger)
	} else {
		logger.Info("connected to analyzer worker", "addr", cfg.AnalyzerAddr)
	}
	defer analysisBackend.Close()

	// Create gRPC server with auth interceptors
	authCfg := auth.Config{Enabled: cfg.AuthEnabled}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.Interceptor(authCfg, logger)),
		grpc.StreamInterceptor(auth.StreamInterceptor(authCfg, logger)),
	)

	// Register engine API
	engineServer := server.NewEngineServer(cfg, logger, db, analysisBackend)
	engine.RegisterEngineAPIServer(grpcServer, engineServer)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("cartomix.engine.EngineAPI", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for grpcurl/grpcui
	reflection.Register(grpcServer)

	// Start listener
	addr := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info("shutting down", "signal", sig)
		healthServer.SetServingStatus("cartomix.engine.EngineAPI", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
	}()

	logger.Info("starting engine server",
		"port", cfg.Port,
		"data_dir", cfg.DataDir,
		"auth_enabled", cfg.AuthEnabled,
	)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
