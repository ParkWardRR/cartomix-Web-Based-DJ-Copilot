package server

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	analyzerpb "github.com/cartomix/cancun/gen/go/analyzer"
	"github.com/cartomix/cancun/gen/go/common"
	eng "github.com/cartomix/cancun/gen/go/engine"
	analyzeriface "github.com/cartomix/cancun/internal/analyzer"
	"github.com/cartomix/cancun/internal/config"
	"github.com/cartomix/cancun/internal/exporter"
	"github.com/cartomix/cancun/internal/planner"
	"github.com/cartomix/cancun/internal/scanner"
	"github.com/cartomix/cancun/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			Path:      p.Path,
			Status:    p.Status,
			Error:     p.Error,
			Processed: p.Processed,
			Total:     p.Total,
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

	for _, track := range tracks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip if already analyzed and not forced
		if !req.GetForce() {
			if rec, err := s.db.LatestAnalysisRecord(track.ID); err == nil && rec.Status == storage.AnalysisStatusComplete && rec.Version >= version {
				stream.Send(&eng.AnalyzeProgress{
					Id:      &common.TrackId{ContentHash: track.ContentHash, Path: track.Path},
					Stage:   "done",
					Status:  "done",
					Percent: 100,
				})
				continue
			}
		}

		if err := stream.Send(&eng.AnalyzeProgress{
			Id:      &common.TrackId{ContentHash: track.ContentHash, Path: track.Path},
			Stage:   "decode",
			Status:  "running",
			Percent: 10,
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
				Id:      job.Id,
				Stage:   "analyze",
				Status:  "error",
				Error:   err.Error(),
				Percent: 100,
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

		if err := stream.Send(&eng.AnalyzeProgress{
			Id:      job.Id,
			Stage:   "done",
			Status:  "done",
			Percent: 100,
		}); err != nil {
			return err
		}
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
