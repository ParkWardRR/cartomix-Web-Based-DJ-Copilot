package config

import (
	"flag"
	"os"
)

type Config struct {
	// Server settings
	Port     int
	DataDir  string
	LogLevel string

	// Analyzer settings
	AnalyzerAddr string

	// Auth settings
	AuthEnabled bool
}

func Parse() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Port, "port", 50051, "gRPC server port")
	flag.StringVar(&cfg.DataDir, "data-dir", defaultDataDir(), "data directory for SQLite and blobs")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "log level (debug, info, warn, error)")
	flag.StringVar(&cfg.AnalyzerAddr, "analyzer-addr", "localhost:50052", "analyzer worker gRPC address")
	flag.BoolVar(&cfg.AuthEnabled, "auth", false, "enable API authentication (default: open for local use)")

	flag.Parse()
	return cfg
}

func defaultDataDir() string {
	if dir := os.Getenv("CARTOMIX_DATA_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cartomix"
	}
	return home + "/.cartomix"
}
