package analyzer

import (
	"context"
	"log/slog"
	"time"

	"github.com/cartomix/cancun/gen/go/analyzer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC analyzer worker client with connection management.
type Client struct {
	conn   *grpc.ClientConn
	client analyzer.AnalyzerWorkerClient
	logger *slog.Logger
}

// NewClient creates a gRPC client for the Swift analyzer worker.
func NewClient(addr string, logger *slog.Logger) (*Client, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: analyzer.NewAnalyzerWorkerClient(conn),
		logger: logger,
	}, nil
}

// AnalyzeTrack sends an analysis job to the Swift analyzer worker.
func (c *Client) AnalyzeTrack(ctx context.Context, job *analyzer.AnalyzeJob) (*analyzer.AnalyzeResult, error) {
	c.logger.Debug("sending analysis job to worker",
		"track_id", job.GetId().GetContentHash(),
		"path", job.GetPath(),
	)

	start := time.Now()
	result, err := c.client.AnalyzeTrack(ctx, job)
	if err != nil {
		c.logger.Error("analysis failed",
			"track_id", job.GetId().GetContentHash(),
			"error", err,
			"duration", time.Since(start),
		)
		return nil, err
	}

	c.logger.Info("analysis complete",
		"track_id", job.GetId().GetContentHash(),
		"duration", time.Since(start),
		"bpm", result.GetAnalysis().GetBeatgrid().GetTempoMap(),
	)

	return result, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
