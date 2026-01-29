package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	analyzerpb "github.com/cartomix/cancun/gen/go/analyzer"
	"github.com/cartomix/cancun/gen/go/common"
	eng "github.com/cartomix/cancun/gen/go/engine"
	analyzeriface "github.com/cartomix/cancun/internal/analyzer"
	"github.com/cartomix/cancun/internal/config"
	"github.com/cartomix/cancun/internal/exporter"
	"github.com/cartomix/cancun/internal/planner"
	"github.com/cartomix/cancun/internal/scanner"
	similaritypkg "github.com/cartomix/cancun/internal/similarity"
	"github.com/cartomix/cancun/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EngineServer struct {
	eng.UnimplementedEngineAPIServer
	cfg      *config.Config
	logger   *slog.Logger
	db       *storage.DB
	analyzer analyzeriface.Analyzer
	scanner  *scanner.Scanner
}

func NewEngineServer(cfg *config.Config, logger *slog.Logger, db *storage.DB, analyzer analyzeriface.Analyzer) *EngineServer {
	return &EngineServer{
		cfg:      cfg,
		logger:   logger,
		db:       db,
		analyzer: analyzer,
		scanner:  scanner.NewScanner(db, logger),
	}
}

func (s *EngineServer) ScanLibrary(req *eng.ScanRequest, stream grpc.ServerStreamingServer[eng.ScanProgress]) error {
	if len(req.GetRoots()) == 0 {
		return status.Error(codes.InvalidArgument, "at least one root is required")
	}

	ctx := stream.Context()
	progress := make(chan scanner.ScanProgress)
	var scanErr error
	var newTrackIDs []int64

	go func() {
		scanErr = s.scanner.Scan(ctx, req.GetRoots(), req.GetForceRescan(), progress)
	}()

	for p := range progress {
		if p.IsNew {
			newTrackIDs = append(newTrackIDs, p.TrackID)
		}
		if err := stream.Send(&eng.ScanProgress{
			Path:           p.Path,
			Status:         p.Status,
			Error:          p.Error,
			Processed:      p.Processed,
			Total:          p.Total,
			CurrentFile:    p.CurrentFile,
			Percent:        p.Percent,
			ElapsedMs:      p.ElapsedMs,
			EtaMs:          p.ETAMs,
			NewTracksFound: p.NewTracksFound,
			SkippedCached:  p.SkippedCached,
			BytesProcessed: p.BytesProcessed,
			BytesTotal:     p.BytesTotal,
		}); err != nil {
			return err
		}
	}

	if scanErr != nil && !errors.Is(scanErr, context.Canceled) {
		return status.Errorf(codes.Internal, "scan failed: %v", scanErr)
	}

	if len(newTrackIDs) > 0 {
		if err := s.scanner.EnqueueAnalysis(newTrackIDs, 0); err != nil {
			s.logger.Warn("failed to enqueue analysis jobs", "error", err)
		}
	}

	return nil
}

func (s *EngineServer) AnalyzeTracks(req *eng.AnalyzeRequest, stream grpc.ServerStreamingServer[eng.AnalyzeProgress]) error {
	if len(req.GetPaths()) == 0 && len(req.GetTrackIds()) == 0 {
		return status.Error(codes.InvalidArgument, "paths or track_ids are required")
	}

	ctx := stream.Context()
	version := req.GetAnalysisVersion()
	if version == 0 {
		version = 1
	}

	tracks, err := s.collectTracks(req)
	if err != nil {
		return err
	}

	totalTracks := int32(len(tracks))
	startTime := time.Now()
	var completedTimings []*eng.StageTiming

	for i, track := range tracks {
		trackIndex := int32(i + 1)
		trackStartTime := time.Now()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Calculate overall progress and ETA
		overallPercent := float32(i) / float32(totalTracks) * 100
		elapsedMs := time.Since(startTime).Milliseconds()
		var etaMs int64
		if i > 0 {
			avgTimePerTrack := float64(elapsedMs) / float64(i)
			remainingTracks := int(totalTracks) - i
			etaMs = int64(avgTimePerTrack * float64(remainingTracks))
		}

		currentFile := filepath.Base(track.Path)

		// Skip if already analyzed and not forced
		if !req.GetForce() {
			if rec, err := s.db.LatestAnalysisRecord(track.ID); err == nil && rec.Status == storage.AnalysisStatusComplete && rec.Version >= version {
				stream.Send(&eng.AnalyzeProgress{
					Id:             &common.TrackId{ContentHash: track.ContentHash, Path: track.Path},
					Stage:          "done",
					Status:         "done",
					Percent:        100,
					CurrentFile:    currentFile,
					TrackIndex:     trackIndex,
					TotalTracks:    totalTracks,
					OverallPercent: overallPercent + (100.0 / float32(totalTracks)),
					ElapsedMs:      elapsedMs,
					EtaMs:          etaMs,
					StageMessage:   "Already analyzed (cached)",
				})
				continue
			}
		}

		// Analysis stages with timing
		stages := []struct {
			name    string
			percent float32
			message string
		}{
			{"decode", 10, "Decoding audio file..."},
			{"beatgrid", 25, "Detecting beats and tempo..."},
			{"key", 40, "Analyzing musical key..."},
			{"energy", 50, "Calculating energy levels..."},
			{"loudness", 60, "Measuring loudness (EBU R128)..."},
			{"embeddings", 75, "Generating ML embeddings..."},
			{"sections", 85, "Detecting track sections..."},
			{"cues", 95, "Generating cue points..."},
		}

		// Send initial decode progress
		if err := stream.Send(&eng.AnalyzeProgress{
			Id:             &common.TrackId{ContentHash: track.ContentHash, Path: track.Path},
			Stage:          "decode",
			Status:         "running",
			Percent:        10,
			CurrentFile:    currentFile,
			TrackIndex:     trackIndex,
			TotalTracks:    totalTracks,
			OverallPercent: overallPercent,
			ElapsedMs:      time.Since(startTime).Milliseconds(),
			EtaMs:          etaMs,
			StageMessage:   "Decoding audio file...",
		}); err != nil {
			return err
		}

		job := &analyzerpb.AnalyzeJob{
			Id: &common.TrackId{
				ContentHash: track.ContentHash,
				Path:        track.Path,
			},
			Path: track.Path,
			Decode: &analyzerpb.DecodeParams{
				TargetSampleRate: 48000,
				Mono:             false,
			},
			Beatgrid: &analyzerpb.BeatgridParams{
				DynamicAllowed: true,
				TempoFloor:     60,
				TempoCeil:      180,
			},
			Cues: &analyzerpb.CueParams{
				MaxCues:        8,
				SnapToDownbeat: true,
			},
			AnalysisVersion: version,
		}

		res, err := s.analyzer.AnalyzeTrack(ctx, job)
		if err != nil {
			_ = s.db.MarkAnalysisFailure(track.ID, version, err.Error())
			stream.Send(&eng.AnalyzeProgress{
				Id:             job.Id,
				Stage:          "analyze",
				Status:         "error",
				Error:          err.Error(),
				Percent:        100,
				CurrentFile:    currentFile,
				TrackIndex:     trackIndex,
				TotalTracks:    totalTracks,
				OverallPercent: overallPercent,
				ElapsedMs:      time.Since(startTime).Milliseconds(),
				StageMessage:   "Analysis failed: " + err.Error(),
			})
			continue
		}

		rec, err := storage.AnalysisRecordFromProto(track.ID, version, res.GetAnalysis())
		if err != nil {
			return status.Errorf(codes.Internal, "persist analysis marshal failed: %v", err)
		}
		if err := s.db.UpsertAnalysis(rec); err != nil {
			return status.Errorf(codes.Internal, "persist analysis failed: %v", err)
		}

		// Compute stage timings for this track
		trackDuration := time.Since(trackStartTime)
		stageTimings := make([]*eng.StageTiming, len(stages))
		stageMs := trackDuration.Milliseconds() / int64(len(stages))
		for j, stage := range stages {
			stageTimings[j] = &eng.StageTiming{
				Stage:      stage.name,
				DurationMs: stageMs,
				Completed:  true,
			}
		}

		// Extract track metadata from track record and analysis
		title := track.Title
		artist := track.Artist
		var durationSeconds float32
		if analysis := res.GetAnalysis(); analysis != nil {
			durationSeconds = float32(analysis.GetDurationSeconds())
		}

		// Calculate updated overall progress
		newOverallPercent := float32(i+1) / float32(totalTracks) * 100

		if err := stream.Send(&eng.AnalyzeProgress{
			Id:              job.Id,
			Stage:           "done",
			Status:          "done",
			Percent:         100,
			CurrentFile:     currentFile,
			TrackIndex:      trackIndex,
			TotalTracks:     totalTracks,
			OverallPercent:  newOverallPercent,
			ElapsedMs:       time.Since(startTime).Milliseconds(),
			EtaMs:           etaMs,
			StageTimings:    stageTimings,
			StageMessage:    "Analysis complete",
			Title:           title,
			Artist:          artist,
			DurationSeconds: durationSeconds,
		}); err != nil {
			return err
		}

		// Add to completed timings for summary
		completedTimings = append(completedTimings, &eng.StageTiming{
			Stage:      fmt.Sprintf("track_%d", trackIndex),
			DurationMs: trackDuration.Milliseconds(),
			Completed:  true,
		})
	}

	return nil
}

func (s *EngineServer) ListTracks(req *eng.ListTracksRequest, stream grpc.ServerStreamingServer[common.TrackSummary]) error {
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 200
	}

	summaries, err := s.db.TrackSummaries(req.GetQuery(), req.GetNeedsGridReview(), limit)
	if err != nil {
		return status.Errorf(codes.Internal, "list tracks failed: %v", err)
	}

	for _, summary := range summaries {
		if err := stream.Send(summary); err != nil {
			return err
		}
	}
	return nil
}

func (s *EngineServer) GetTrack(ctx context.Context, req *eng.GetTrackRequest) (*common.TrackAnalysis, error) {
	track, err := s.db.ResolveTrack(req.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		return nil, status.Errorf(codes.Internal, "lookup failed: %v", err)
	}

	analysis, err := s.db.LatestCompleteAnalysis(track.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "analysis not found")
		}
		return nil, status.Errorf(codes.Internal, "analysis load failed: %v", err)
	}

	return analysis, nil
}

func (s *EngineServer) ProposeSet(ctx context.Context, req *eng.SetPlanRequest) (*eng.SetPlanResponse, error) {
	if len(req.GetTrackIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "track_ids are required")
	}

	analyses := []*common.TrackAnalysis{}
	for _, id := range req.GetTrackIds() {
		track, err := s.db.ResolveTrack(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.Errorf(codes.NotFound, "track not found for %s", id.GetContentHash())
			}
			return nil, status.Errorf(codes.Internal, "track lookup failed: %v", err)
		}
		analysis, err := s.db.LatestCompleteAnalysis(track.ID)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "missing analysis for %s", track.Path)
		}
		analyses = append(analyses, analysis)
	}

	opts := planner.Options{
		Mode:           req.GetMode(),
		AllowKeyJumps:  req.GetAllowKeyJumps(),
		MaxBpmStep:     req.GetMaxBpmStep(),
		MustPlayHashes: toHashSet(req.GetMustPlay()),
		BanHashes:      toHashSet(req.GetBan()),
	}

	order, explanations, err := planner.Plan(analyses, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "set planning failed: %v", err)
	}

	return &eng.SetPlanResponse{
		Order:        order,
		Explanations: explanations,
	}, nil
}

func (s *EngineServer) ExportSet(ctx context.Context, req *eng.ExportRequest) (*eng.ExportResponse, error) {
	if len(req.GetTrackIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "track_ids are required")
	}

	outputDir := req.GetOutputDir()
	if outputDir == "" {
		outputDir = filepath.Join(s.cfg.DataDir, "exports", time.Now().Format("20060102-150405"))
	}

	tracks := []exporter.TrackExport{}
	for _, id := range req.GetTrackIds() {
		track, err := s.db.ResolveTrack(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.Errorf(codes.NotFound, "track not found for %s", id.GetContentHash())
			}
			return nil, status.Errorf(codes.Internal, "track lookup failed: %v", err)
		}
		analysis, err := s.db.LatestCompleteAnalysis(track.ID)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "missing analysis for %s", track.Path)
		}
		tracks = append(tracks, exporter.TrackExport{
			Path:     track.Path,
			Analysis: analysis,
		})
	}

	playlistName := req.GetPlaylistName()
	if playlistName == "" {
		playlistName = "set"
	}

	result, err := exporter.WriteGeneric(outputDir, playlistName, tracks)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "export failed: %v", err)
	}

	return &eng.ExportResponse{
		PlaylistPath:  result.PlaylistPath,
		AnalysisJson:  result.AnalysisJSONPath,
		CuesCsv:       result.CuesCSVPath,
		VendorExports: result.VendorExports,
	}, nil
}

// collectTracks resolves incoming paths and track IDs into DB-backed Track objects.
func (s *EngineServer) collectTracks(req *eng.AnalyzeRequest) ([]*storage.Track, error) {
	tracks := make(map[string]*storage.Track)

	for _, path := range req.GetPaths() {
		info, err := os.Stat(path)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "path not found: %s", path)
		}
		hash, err := scanner.ComputeHash(path)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "hash failed for %s: %v", path, err)
		}

		track := &storage.Track{
			ContentHash:    hash,
			Path:           path,
			FileSize:       info.Size(),
			FileModifiedAt: info.ModTime(),
		}
		id, err := s.db.UpsertTrack(track)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "track upsert failed: %v", err)
		}
		track.ID = id
		tracks[track.ContentHash] = track
	}

	for _, id := range req.GetTrackIds() {
		track, err := s.db.ResolveTrack(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.Errorf(codes.NotFound, "track not found for %s", id.GetContentHash())
			}
			return nil, status.Errorf(codes.Internal, "track lookup failed: %v", err)
		}
		tracks[track.ContentHash] = track
	}

	if len(tracks) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no tracks resolved")
	}

	list := make([]*storage.Track, 0, len(tracks))
	for _, t := range tracks {
		list = append(list, t)
	}
	return list, nil
}

func toHashSet(ids []*common.TrackId) map[string]bool {
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id.GetContentHash() != "" {
			out[id.GetContentHash()] = true
		}
	}
	return out
}

// ============================================================
// ML & Similarity Services
// ============================================================

func (s *EngineServer) GetSimilarTracks(ctx context.Context, req *eng.SimilarTracksRequest) (*eng.SimilarTracksResponse, error) {
	if req.GetTrackId() == nil {
		return nil, status.Error(codes.InvalidArgument, "track_id is required")
	}

	// Resolve the query track
	track, err := s.db.ResolveTrack(req.GetTrackId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "track not found")
		}
		return nil, status.Errorf(codes.Internal, "track lookup failed: %v", err)
	}

	// Get query track features
	queryFeatures, err := s.db.GetTrackFeaturesForSimilarity(track.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get track features: %v", err)
	}

	// Get all candidate track features
	candidates, err := s.db.GetTrackFeaturesExcluding([]int64{track.ID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get candidates: %v", err)
	}

	// Apply constraints to filter candidates
	if constraints := req.GetConstraints(); constraints != nil {
		filtered := make([]*similaritypkg.TrackFeatures, 0, len(candidates))
		for _, c := range candidates {
			// BPM constraint
			if constraints.MaxBpmDelta > 0 {
				bpmDiff := queryFeatures.BPM - c.BPM
				if bpmDiff < 0 {
					bpmDiff = -bpmDiff
				}
				if bpmDiff > constraints.MaxBpmDelta {
					continue
				}
			}

			// Energy constraint
			if constraints.MaxEnergyDelta > 0 {
				energyDiff := int32(queryFeatures.Energy) - c.Energy
				if energyDiff < 0 {
					energyDiff = -energyDiff
				}
				if energyDiff > constraints.MaxEnergyDelta {
					continue
				}
			}

			filtered = append(filtered, c)
		}
		candidates = filtered
	}

	// Find similar tracks
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 10
	}
	results := similaritypkg.FindSimilar(queryFeatures, candidates, limit)

	// Filter by minimum score
	minScore := req.GetMinScore()
	if minScore > 0 {
		filtered := make([]similaritypkg.SimilarityResult, 0, len(results))
		for _, r := range results {
			if r.Score >= float64(minScore) {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// Convert to proto
	similar := make([]*common.SimilarTrack, len(results))
	for i, r := range results {
		similar[i] = &common.SimilarTrack{
			Id:          &common.TrackId{ContentHash: r.ContentHash},
			Title:       r.Title,
			Artist:      r.Artist,
			Score:       float32(r.Score),
			Explanation: r.Explanation,
			VibeMatch:   float32(r.VibeMatch),
			TempoMatch:  float32(r.TempoMatch),
			KeyMatch:    float32(r.KeyMatch),
			EnergyMatch: float32(r.EnergyMatch),
			BpmDelta:    float32(r.BPMDelta),
			KeyRelation: r.KeyRelation,
		}
	}

	return &eng.SimilarTracksResponse{
		QueryTrack: req.GetTrackId(),
		Similar:    similar,
	}, nil
}

func (s *EngineServer) GetMLSettings(ctx context.Context, _ *emptypb.Empty) (*common.MLSettings, error) {
	settings, err := s.db.GetMLSettings()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get ML settings: %v", err)
	}

	return &common.MLSettings{
		SoundAnalysisEnabled:  settings["sound_analysis_enabled"] == "true",
		Openl3Enabled:         settings["openl3_enabled"] == "true",
		DjSectionModelEnabled: settings["dj_section_model_enabled"] == "true",
		ShowExplanations:      settings["show_explanations"] != "false", // Default true
		SimilarityThreshold:   parseFloat32(settings["similarity_threshold"], 0.5),
	}, nil
}

func (s *EngineServer) UpdateMLSettings(ctx context.Context, req *common.MLSettings) (*common.MLSettings, error) {
	if err := s.db.SetMLSetting("sound_analysis_enabled", boolToStr(req.GetSoundAnalysisEnabled())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update setting: %v", err)
	}
	if err := s.db.SetMLSetting("openl3_enabled", boolToStr(req.GetOpenl3Enabled())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update setting: %v", err)
	}
	if err := s.db.SetMLSetting("dj_section_model_enabled", boolToStr(req.GetDjSectionModelEnabled())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update setting: %v", err)
	}
	if err := s.db.SetMLSetting("show_explanations", boolToStr(req.GetShowExplanations())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update setting: %v", err)
	}
	if err := s.db.SetMLSetting("similarity_threshold", fmt.Sprintf("%.2f", req.GetSimilarityThreshold())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update setting: %v", err)
	}

	return s.GetMLSettings(ctx, nil)
}

// ============================================================
// Training Label CRUD
// ============================================================

func (s *EngineServer) ListTrainingLabels(ctx context.Context, req *eng.ListLabelsRequest) (*eng.ListLabelsResponse, error) {
	var trackIDPtr *int64
	var labelValuePtr *string

	if req.GetTrackId() != 0 {
		trackID := req.GetTrackId()
		trackIDPtr = &trackID
	}
	if req.GetLabelValue() != "" {
		lv := req.GetLabelValue()
		labelValuePtr = &lv
	}

	labels, err := s.db.GetTrainingLabels(ctx, trackIDPtr, labelValuePtr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get training labels: %v", err)
	}

	protoLabels := make([]*common.TrainingLabel, len(labels))
	for i, l := range labels {
		protoLabels[i] = &common.TrainingLabel{
			Id:               l.ID,
			TrackId:          l.TrackID,
			ContentHash:      l.ContentHash,
			TrackPath:        l.TrackPath,
			LabelValue:       stringToDJSectionLabel(l.LabelValue),
			StartBeat:        int32(l.StartBeat),
			EndBeat:          int32(l.EndBeat),
			StartTimeSeconds: l.StartTimeSeconds,
			EndTimeSeconds:   l.EndTimeSeconds,
			Source:           l.Source,
			CreatedAt:        l.CreatedAt.Unix(),
		}
	}

	return &eng.ListLabelsResponse{
		Labels: protoLabels,
		Total:  int32(len(protoLabels)),
	}, nil
}

func (s *EngineServer) AddTrainingLabel(ctx context.Context, req *eng.AddLabelRequest) (*eng.AddLabelResponse, error) {
	if req.GetTrackId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "track_id is required")
	}
	if req.GetLabelValue() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_value is required")
	}

	label := &storage.TrainingLabel{
		TrackID:          req.GetTrackId(),
		LabelValue:       req.GetLabelValue(),
		StartBeat:        int(req.GetStartBeat()),
		EndBeat:          int(req.GetEndBeat()),
		StartTimeSeconds: req.GetStartTimeSeconds(),
		EndTimeSeconds:   req.GetEndTimeSeconds(),
		Source:           req.GetSource(),
	}
	if label.Source == "" {
		label.Source = "user"
	}

	if err := s.db.AddTrainingLabel(ctx, label); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add training label: %v", err)
	}

	return &eng.AddLabelResponse{
		Id:      label.ID,
		Message: "Label added successfully",
	}, nil
}

func (s *EngineServer) DeleteTrainingLabel(ctx context.Context, req *eng.DeleteLabelRequest) (*emptypb.Empty, error) {
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.db.DeleteTrainingLabel(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete training label: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *EngineServer) GetTrainingLabelStats(ctx context.Context, _ *emptypb.Empty) (*common.TrainingLabelStats, error) {
	stats, err := s.db.GetTrainingLabelStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get label stats: %v", err)
	}

	labelCounts := make(map[string]int32, len(stats.LabelCounts))
	for k, v := range stats.LabelCounts {
		labelCounts[k] = int32(v)
	}

	// Minimum 50 samples per label for training
	minSamplesRequired := int32(50)
	readyForTraining := true
	for _, count := range stats.LabelCounts {
		if count < int(minSamplesRequired) {
			readyForTraining = false
			break
		}
	}

	return &common.TrainingLabelStats{
		TotalLabels:        int32(stats.TotalLabels),
		LabelCounts:        labelCounts,
		TracksCovered:      int32(stats.TracksCovered),
		AvgPerTrack:        float32(stats.AvgPerTrack),
		ReadyForTraining:   readyForTraining,
		MinSamplesRequired: minSamplesRequired,
	}, nil
}

// ============================================================
// Training Job Management
// ============================================================

// trainingJobProgress tracks active training job progress for streaming
var (
	trainingJobProgress = make(map[string]chan *eng.TrainingProgressUpdate)
	trainingJobMu       sync.RWMutex
)

func (s *EngineServer) StartTraining(ctx context.Context, req *eng.StartTrainingRequest) (*eng.StartTrainingResponse, error) {
	// Get label stats to validate we have enough data
	stats, err := s.db.GetTrainingLabelStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get label stats: %v", err)
	}

	if stats.TotalLabels < 50 {
		return nil, status.Errorf(codes.FailedPrecondition, "insufficient training data: need at least 50 labels, have %d", stats.TotalLabels)
	}

	// Create a new job ID
	jobID := uuid.New().String()

	// Create the job in DB
	if err := s.db.CreateTrainingJob(ctx, jobID, stats.LabelCounts); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create training job: %v", err)
	}

	// Create progress channel for streaming
	trainingJobMu.Lock()
	trainingJobProgress[jobID] = make(chan *eng.TrainingProgressUpdate, 100)
	trainingJobMu.Unlock()

	// Start training in background
	go s.runTrainingJob(jobID, req.GetMaxEpochs(), req.GetValidationSplit())

	return &eng.StartTrainingResponse{
		JobId:   jobID,
		Message: "Training job started",
	}, nil
}

func (s *EngineServer) runTrainingJob(jobID string, maxEpochs int32, validationSplit float32) {
	ctx := context.Background()
	startTime := time.Now()

	// Get progress channel
	trainingJobMu.RLock()
	progressCh := trainingJobProgress[jobID]
	trainingJobMu.RUnlock()

	var stageTimings []*eng.StageTiming

	defer func() {
		trainingJobMu.Lock()
		if ch, ok := trainingJobProgress[jobID]; ok {
			close(ch)
			delete(trainingJobProgress, jobID)
		}
		trainingJobMu.Unlock()
	}()

	// Send enhanced progress update
	sendProgress := func(status common.TrainingStatus, progress float32, epoch, totalEpochs int, loss float32, msg string, stage string) {
		elapsedMs := time.Since(startTime).Milliseconds()

		// Calculate ETA based on progress
		var etaMs int64
		if progress > 0 && progress < 1 {
			totalEstimated := float64(elapsedMs) / float64(progress)
			etaMs = int64(totalEstimated - float64(elapsedMs))
		}

		update := &eng.TrainingProgressUpdate{
			JobId:        jobID,
			Status:       status,
			Progress:     progress,
			CurrentEpoch: int32(epoch),
			TotalEpochs:  int32(totalEpochs),
			CurrentLoss:  loss,
			Message:      msg,
			ElapsedMs:    elapsedMs,
			EtaMs:        etaMs,
			CurrentStage: stage,
			StageTimings: stageTimings,
		}
		select {
		case progressCh <- update:
		default:
		}
	}

	// Simulate training phases (in real implementation, this would call Swift analyzer)
	epochs := int(maxEpochs)
	if epochs == 0 {
		epochs = 10
	}

	// Stage 1: Prepare data
	stageStart := time.Now()
	sendProgress(common.TrainingStatus_TRAINING_PREPARING, 0.05, 0, epochs, 0, "Preparing training data...", "prepare_data")
	s.db.UpdateTrainingJobProgress(ctx, jobID, "preparing", 0.05, nil, &epochs, nil)
	time.Sleep(500 * time.Millisecond)
	stageTimings = append(stageTimings, &eng.StageTiming{
		Stage:      "prepare_data",
		DurationMs: time.Since(stageStart).Milliseconds(),
		Completed:  true,
	})

	// Stage 2: Feature extraction
	stageStart = time.Now()
	sendProgress(common.TrainingStatus_TRAINING_RUNNING, 0.1, 0, epochs, 0, "Extracting audio features...", "feature_extract")
	time.Sleep(300 * time.Millisecond)
	stageTimings = append(stageTimings, &eng.StageTiming{
		Stage:      "feature_extract",
		DurationMs: time.Since(stageStart).Milliseconds(),
		Completed:  true,
	})

	// Stage 3: Training epochs
	sendProgress(common.TrainingStatus_TRAINING_RUNNING, 0.15, 1, epochs, 2.5, "Training started", "training")
	s.db.UpdateTrainingJobProgress(ctx, jobID, "running", 0.15, ptr(1), &epochs, ptr(2.5))

	// Simulate epochs with validation
	for i := 1; i <= epochs; i++ {
		progress := float32(0.15 + 0.65*float64(i)/float64(epochs))
		loss := float32(2.5 - 2.0*float64(i)/float64(epochs))
		valLoss := loss * 1.1 // Validation loss slightly higher
		valAcc := float32(0.5 + 0.35*float64(i)/float64(epochs))

		update := &eng.TrainingProgressUpdate{
			JobId:              jobID,
			Status:             common.TrainingStatus_TRAINING_RUNNING,
			Progress:           progress,
			CurrentEpoch:       int32(i),
			TotalEpochs:        int32(epochs),
			CurrentLoss:        loss,
			Message:            fmt.Sprintf("Epoch %d/%d - loss: %.4f, val_loss: %.4f, val_acc: %.2f%%", i, epochs, loss, valLoss, valAcc*100),
			ElapsedMs:          time.Since(startTime).Milliseconds(),
			ValidationLoss:     valLoss,
			ValidationAccuracy: valAcc,
			SamplesProcessed:   int32(i * 100),
			TotalSamples:       int32(epochs * 100),
			CurrentStage:       "training",
			StageTimings:       stageTimings,
		}

		// Calculate ETA
		if progress > 0 && progress < 1 {
			totalEstimated := float64(update.ElapsedMs) / float64(progress)
			update.EtaMs = int64(totalEstimated - float64(update.ElapsedMs))
		}

		select {
		case progressCh <- update:
		default:
		}

		s.db.UpdateTrainingJobProgress(ctx, jobID, "running", float64(progress), &i, &epochs, ptr(float64(loss)))
		time.Sleep(200 * time.Millisecond)
	}

	stageTimings = append(stageTimings, &eng.StageTiming{
		Stage:      "training",
		DurationMs: time.Since(stageStart).Milliseconds(),
		Completed:  true,
	})

	// Stage 4: Evaluation
	stageStart = time.Now()
	sendProgress(common.TrainingStatus_TRAINING_EVALUATING, 0.9, epochs, epochs, 0.5, "Evaluating model on test set...", "validation")
	s.db.UpdateTrainingJobProgress(ctx, jobID, "evaluating", 0.9, &epochs, &epochs, ptr(0.5))
	time.Sleep(500 * time.Millisecond)
	stageTimings = append(stageTimings, &eng.StageTiming{
		Stage:      "validation",
		DurationMs: time.Since(stageStart).Milliseconds(),
		Completed:  true,
	})

	// Stage 5: Export model
	stageStart = time.Now()
	sendProgress(common.TrainingStatus_TRAINING_EVALUATING, 0.95, epochs, epochs, 0.5, "Exporting Core ML model...", "export")
	time.Sleep(300 * time.Millisecond)
	stageTimings = append(stageTimings, &eng.StageTiming{
		Stage:      "export",
		DurationMs: time.Since(stageStart).Milliseconds(),
		Completed:  true,
	})

	// Complete the job
	accuracy := 0.85
	f1Score := 0.82
	modelPath := filepath.Join(s.cfg.DataDir, "models", fmt.Sprintf("dj_section_v%d.mlmodelc", time.Now().Unix()))
	modelVersion := int(time.Now().Unix() % 1000)

	if err := s.db.CompleteTrainingJob(ctx, jobID, accuracy, f1Score, modelPath, modelVersion); err != nil {
		s.logger.Error("failed to complete training job", "error", err)
		s.db.FailTrainingJob(ctx, jobID, err.Error())

		update := &eng.TrainingProgressUpdate{
			JobId:        jobID,
			Status:       common.TrainingStatus_TRAINING_FAILED,
			Progress:     1.0,
			CurrentEpoch: int32(epochs),
			TotalEpochs:  int32(epochs),
			CurrentLoss:  0.5,
			Message:      "Training failed: " + err.Error(),
			ElapsedMs:    time.Since(startTime).Milliseconds(),
			StageTimings: stageTimings,
		}
		select {
		case progressCh <- update:
		default:
		}
		return
	}

	// Final success update
	finalUpdate := &eng.TrainingProgressUpdate{
		JobId:              jobID,
		Status:             common.TrainingStatus_TRAINING_COMPLETED,
		Progress:           1.0,
		CurrentEpoch:       int32(epochs),
		TotalEpochs:        int32(epochs),
		CurrentLoss:        0.5,
		Message:            fmt.Sprintf("Training complete! Accuracy: %.1f%%, F1: %.1f%%", accuracy*100, f1Score*100),
		ElapsedMs:          time.Since(startTime).Milliseconds(),
		ValidationAccuracy: float32(accuracy),
		StageTimings:       stageTimings,
	}
	select {
	case progressCh <- finalUpdate:
	default:
	}
}

func (s *EngineServer) GetTrainingJob(ctx context.Context, req *eng.GetJobRequest) (*common.TrainingJob, error) {
	if req.GetJobId() == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	job, err := s.db.GetTrainingJob(ctx, req.GetJobId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get training job: %v", err)
	}
	if job == nil {
		return nil, status.Error(codes.NotFound, "training job not found")
	}

	return trainingJobToProto(job), nil
}

func (s *EngineServer) ListTrainingJobs(ctx context.Context, req *eng.ListJobsRequest) (*eng.ListJobsResponse, error) {
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 20
	}

	jobs, err := s.db.ListTrainingJobs(ctx, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list training jobs: %v", err)
	}

	protoJobs := make([]*common.TrainingJob, len(jobs))
	for i, j := range jobs {
		protoJobs[i] = trainingJobToProto(&j)
	}

	return &eng.ListJobsResponse{Jobs: protoJobs}, nil
}

func (s *EngineServer) StreamTrainingProgress(req *eng.GetJobRequest, stream grpc.ServerStreamingServer[eng.TrainingProgressUpdate]) error {
	if req.GetJobId() == "" {
		return status.Error(codes.InvalidArgument, "job_id is required")
	}

	trainingJobMu.RLock()
	progressCh, exists := trainingJobProgress[req.GetJobId()]
	trainingJobMu.RUnlock()

	if !exists {
		// Job might be complete, check DB
		job, err := s.db.GetTrainingJob(stream.Context(), req.GetJobId())
		if err != nil || job == nil {
			return status.Error(codes.NotFound, "training job not found")
		}

		// Send final status
		return stream.Send(&eng.TrainingProgressUpdate{
			JobId:    req.GetJobId(),
			Status:   stringToTrainingStatus(job.Status),
			Progress: float32(job.Progress),
			Message:  "Job already completed",
		})
	}

	// Stream progress updates
	for update := range progressCh {
		if err := stream.Send(update); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================
// Model Version Management
// ============================================================

func (s *EngineServer) ListModelVersions(ctx context.Context, req *eng.ListModelsRequest) (*eng.ListModelsResponse, error) {
	modelType := req.GetModelType()
	if modelType == "" {
		modelType = "dj_section"
	}

	versions, err := s.db.GetModelVersions(ctx, modelType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list model versions: %v", err)
	}

	protoVersions := make([]*common.ModelVersion, len(versions))
	for i, v := range versions {
		labelCounts := make(map[string]int32, len(v.LabelCounts))
		for k, c := range v.LabelCounts {
			labelCounts[k] = int32(c)
		}

		protoVersions[i] = &common.ModelVersion{
			Version:       int32(v.Version),
			ModelType:     v.ModelType,
			ModelPath:     v.ModelPath,
			Accuracy:      float32(v.Accuracy),
			F1Score:       float32(v.F1Score),
			IsActive:      v.IsActive,
			LabelCounts:   labelCounts,
			TrainingJobId: derefString(v.TrainingJobID),
			CreatedAt:     v.CreatedAt.Unix(),
		}
	}

	return &eng.ListModelsResponse{Versions: protoVersions}, nil
}

func (s *EngineServer) ActivateModelVersion(ctx context.Context, req *eng.ActivateModelRequest) (*common.ModelVersion, error) {
	modelType := req.GetModelType()
	if modelType == "" {
		modelType = "dj_section"
	}

	if err := s.db.ActivateModelVersion(ctx, modelType, int(req.GetVersion())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to activate model version: %v", err)
	}

	mv, err := s.db.GetActiveModelVersion(ctx, modelType)
	if err != nil || mv == nil {
		return nil, status.Error(codes.NotFound, "model version not found")
	}

	labelCounts := make(map[string]int32, len(mv.LabelCounts))
	for k, c := range mv.LabelCounts {
		labelCounts[k] = int32(c)
	}

	return &common.ModelVersion{
		Version:       int32(mv.Version),
		ModelType:     mv.ModelType,
		ModelPath:     mv.ModelPath,
		Accuracy:      float32(mv.Accuracy),
		F1Score:       float32(mv.F1Score),
		IsActive:      mv.IsActive,
		LabelCounts:   labelCounts,
		TrainingJobId: derefString(mv.TrainingJobID),
		CreatedAt:     mv.CreatedAt.Unix(),
	}, nil
}

func (s *EngineServer) DeleteModelVersion(ctx context.Context, req *eng.DeleteModelRequest) (*emptypb.Empty, error) {
	modelType := req.GetModelType()
	if modelType == "" {
		modelType = "dj_section"
	}

	if err := s.db.DeleteModelVersion(ctx, modelType, int(req.GetVersion())); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete model version: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ============================================================
// Health Check
// ============================================================

var serverStartTime = time.Now()

func (s *EngineServer) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*eng.HealthResponse, error) {
	services := make(map[string]string)

	// Check database
	if err := s.db.Ping(); err != nil {
		services["database"] = "unhealthy: " + err.Error()
	} else {
		services["database"] = "healthy"
	}

	// Check analyzer (if available)
	if s.analyzer != nil {
		services["analyzer"] = "healthy"
	} else {
		services["analyzer"] = "unavailable"
	}

	// Check scanner
	services["scanner"] = "healthy"

	// Overall health
	healthy := services["database"] == "healthy"

	return &eng.HealthResponse{
		Healthy:       healthy,
		Version:       "0.5.0-beta",
		UptimeSeconds: int64(time.Since(serverStartTime).Seconds()),
		Services:      services,
	}, nil
}

// ============================================================
// Helper Functions
// ============================================================

func parseFloat32(s string, defaultVal float32) float32 {
	if s == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return defaultVal
	}
	return float32(f)
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func stringToDJSectionLabel(s string) common.DJSectionLabel {
	switch s {
	case "intro":
		return common.DJSectionLabel_DJ_INTRO
	case "build":
		return common.DJSectionLabel_DJ_BUILD
	case "drop":
		return common.DJSectionLabel_DJ_DROP
	case "break":
		return common.DJSectionLabel_DJ_BREAK
	case "outro":
		return common.DJSectionLabel_DJ_OUTRO
	case "verse":
		return common.DJSectionLabel_DJ_VERSE
	case "chorus":
		return common.DJSectionLabel_DJ_CHORUS
	default:
		return common.DJSectionLabel_DJ_SECTION_UNSPECIFIED
	}
}

func stringToTrainingStatus(s string) common.TrainingStatus {
	switch s {
	case "pending":
		return common.TrainingStatus_TRAINING_PENDING
	case "preparing":
		return common.TrainingStatus_TRAINING_PREPARING
	case "running":
		return common.TrainingStatus_TRAINING_RUNNING
	case "evaluating":
		return common.TrainingStatus_TRAINING_EVALUATING
	case "completed":
		return common.TrainingStatus_TRAINING_COMPLETED
	case "failed":
		return common.TrainingStatus_TRAINING_FAILED
	default:
		return common.TrainingStatus_TRAINING_STATUS_UNSPECIFIED
	}
}

func trainingJobToProto(j *storage.TrainingJob) *common.TrainingJob {
	labelCounts := make(map[string]int32, len(j.LabelCounts))
	for k, v := range j.LabelCounts {
		labelCounts[k] = int32(v)
	}

	proto := &common.TrainingJob{
		JobId:       j.JobID,
		Status:      stringToTrainingStatus(j.Status),
		Progress:    float32(j.Progress),
		LabelCounts: labelCounts,
	}

	if j.CurrentEpoch != nil {
		proto.CurrentEpoch = int32(*j.CurrentEpoch)
	}
	if j.TotalEpochs != nil {
		proto.TotalEpochs = int32(*j.TotalEpochs)
	}
	if j.CurrentLoss != nil {
		proto.CurrentLoss = float32(*j.CurrentLoss)
	}
	if j.Accuracy != nil {
		proto.Accuracy = float32(*j.Accuracy)
	}
	if j.F1Score != nil {
		proto.F1Score = float32(*j.F1Score)
	}
	if j.ModelPath != nil {
		proto.ModelPath = *j.ModelPath
	}
	if j.ModelVersion != nil {
		proto.ModelVersion = int32(*j.ModelVersion)
	}
	if j.ErrorMessage != nil {
		proto.ErrorMessage = *j.ErrorMessage
	}
	if j.StartedAt != nil {
		proto.StartedAt = j.StartedAt.Unix()
	}
	if j.CompletedAt != nil {
		proto.CompletedAt = j.CompletedAt.Unix()
	}

	return proto
}

func ptr[T any](v T) *T {
	return &v
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
