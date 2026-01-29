package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/cartomix/cancun/gen/go/analyzer"
	"github.com/cartomix/cancun/gen/go/common"
	"google.golang.org/protobuf/types/known/durationpb"
)

// CPUFallback provides a basic CPU-based analyzer for development and
// systems without the Swift analyzer. It produces placeholder results
// with basic file hashing but no actual audio analysis.
type CPUFallback struct {
	logger *slog.Logger
}

// NewCPUFallback creates a new CPU fallback analyzer.
func NewCPUFallback(logger *slog.Logger) *CPUFallback {
	return &CPUFallback{logger: logger}
}

// AnalyzeTrack produces stub analysis results for testing the pipeline.
// Real analysis requires the Swift analyzer with Accelerate/CoreML.
func (f *CPUFallback) AnalyzeTrack(ctx context.Context, job *analyzer.AnalyzeJob) (*analyzer.AnalyzeResult, error) {
	f.logger.Warn("using CPU fallback analyzer - results are placeholders",
		"path", job.GetPath(),
	)

	// Compute content hash from file
	contentHash, err := hashFile(job.GetPath())
	if err != nil {
		return nil, fmt.Errorf("failed to hash file: %w", err)
	}

	// Create placeholder analysis
	analysis := &common.TrackAnalysis{
		Id: &common.TrackId{
			ContentHash: contentHash,
			Path:        job.GetPath(),
		},
		DurationSeconds: 180.0, // placeholder 3 minutes
		Beatgrid: &common.Beatgrid{
			Beats: generatePlaceholderBeats(128.0, 180.0),
			TempoMap: []*common.TempoMapNode{
				{BeatIndex: 0, Bpm: 128.0},
			},
			Confidence: 0.0, // zero confidence indicates fallback
			IsDynamic:  false,
		},
		Key: &common.MusicalKey{
			Value:      "8A",
			Format:     common.KeyFormat_CAMELOT,
			Confidence: 0.0,
		},
		EnergyGlobal: 5,
		EnergySegments: []*common.EnergySegment{
			{StartBeat: 0, EndBeat: 384, Level: 5},
		},
		Sections: []*common.Section{
			{StartBeat: 0, EndBeat: 64, Label: common.SectionLabel_INTRO, Confidence: 0.0},
			{StartBeat: 64, EndBeat: 320, Label: common.SectionLabel_DROP, Confidence: 0.0},
			{StartBeat: 320, EndBeat: 384, Label: common.SectionLabel_OUTRO, Confidence: 0.0},
		},
		CuePoints: []*common.CuePoint{
			{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_LOAD, Confidence: 0.0},
			{BeatIndex: 64, Time: durationpb.New(30 * time.Second), Type: common.CueType_CUE_DROP, Confidence: 0.0},
		},
		TransitionWindows: []*common.TransitionWindow{
			{StartBeat: 0, EndBeat: 32, Tag: "intro_mix", Confidence: 0.0},
			{StartBeat: 352, EndBeat: 384, Tag: "outro_mix", Confidence: 0.0},
		},
		Loudness: &common.Loudness{
			IntegratedLufs: -14.0,
			TruePeakDb:     -1.0,
		},
		AnalysisVersion: 0, // version 0 indicates fallback
	}

	return &analyzer.AnalyzeResult{
		Analysis: analysis,
	}, nil
}

// Close is a no-op for the CPU fallback.
func (f *CPUFallback) Close() error {
	return nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	// Only hash first 64KB for speed - content hash is just for identity
	_, err = io.CopyN(h, file, 64*1024)
	if err != nil && err != io.EOF {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func generatePlaceholderBeats(bpm, durationSec float64) []*common.BeatMarker {
	beatIntervalSec := 60.0 / bpm
	numBeats := int(durationSec / beatIntervalSec)

	beats := make([]*common.BeatMarker, numBeats)
	for i := 0; i < numBeats; i++ {
		timeSec := float64(i) * beatIntervalSec
		beats[i] = &common.BeatMarker{
			Index:      int32(i),
			Time:       durationpb.New(time.Duration(timeSec * float64(time.Second))),
			IsDownbeat: i%4 == 0,
		}
	}

	return beats
}
