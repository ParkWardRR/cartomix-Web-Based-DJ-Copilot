package analyzer

import (
	"context"

	"github.com/cartomix/cancun/gen/go/analyzer"
)

// Analyzer abstracts the analysis backend - can be remote gRPC or local CPU fallback.
type Analyzer interface {
	AnalyzeTrack(ctx context.Context, job *analyzer.AnalyzeJob) (*analyzer.AnalyzeResult, error)
	Close() error
}
