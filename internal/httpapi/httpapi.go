package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	analyzerpb "github.com/cartomix/cancun/gen/go/analyzer"
	"github.com/cartomix/cancun/gen/go/common"
	"github.com/cartomix/cancun/gen/go/engine"
	"github.com/cartomix/cancun/internal/analyzer"
	"github.com/cartomix/cancun/internal/config"
	"github.com/cartomix/cancun/internal/exporter"
	"github.com/cartomix/cancun/internal/planner"
	"github.com/cartomix/cancun/internal/scanner"
	"github.com/cartomix/cancun/internal/storage"
)

// Server provides HTTP REST endpoints for the Algiers engine.
type Server struct {
	cfg      *config.Config
	logger   *slog.Logger
	db       *storage.DB
	analyzer analyzer.Analyzer
	scanner  *scanner.Scanner
	mux      *http.ServeMux
}

// NewServer creates a new HTTP API server.
func NewServer(cfg *config.Config, logger *slog.Logger, db *storage.DB, az analyzer.Analyzer) *Server {
	s := &Server{
		cfg:      cfg,
		logger:   logger,
		db:       db,
		analyzer: az,
		scanner:  scanner.NewScanner(db, logger),
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return corsMiddleware(s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/tracks", s.handleListTracks)
	s.mux.HandleFunc("GET /api/tracks/{id}", s.handleGetTrack)
	s.mux.HandleFunc("POST /api/scan", s.handleScan)
	s.mux.HandleFunc("POST /api/analyze", s.handleAnalyze)
	s.mux.HandleFunc("POST /api/set/propose", s.handleProposeSet)
	s.mux.HandleFunc("POST /api/export", s.handleExport)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// TrackSummaryResponse is the JSON response for track listings.
type TrackSummaryResponse struct {
	ID           int64   `json:"id"`
	ContentHash  string  `json:"content_hash"`
	Path         string  `json:"path"`
	Title        string  `json:"title"`
	Artist       string  `json:"artist"`
	BPM          float64 `json:"bpm"`
	Key          string  `json:"key"`
	Energy       int32   `json:"energy"`
	CueCount     int32   `json:"cue_count"`
	Status       string  `json:"status"`
	NeedsReview  bool    `json:"needs_review"`
	AnalyzedAt   string  `json:"analyzed_at,omitempty"`
}

func (s *Server) handleListTracks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	needsReview := r.URL.Query().Get("needs_review") == "true"
	limitStr := r.URL.Query().Get("limit")
	limit := 200
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	summaries, err := s.db.TrackSummaries(query, needsReview, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tracks: "+err.Error())
		return
	}

	response := make([]TrackSummaryResponse, 0, len(summaries))
	for _, sum := range summaries {
		keyStr := ""
		if k := sum.GetKey(); k != nil {
			keyStr = k.GetValue()
		}
		response = append(response, TrackSummaryResponse{
			ContentHash: sum.GetId().GetContentHash(),
			Path:        sum.GetId().GetPath(),
			Title:       sum.GetTitle(),
			Artist:      sum.GetArtist(),
			BPM:         sum.GetBpm(),
			Key:         keyStr,
			Energy:      sum.GetEnergy(),
			CueCount:    sum.GetCueCount(),
			Status:      sum.GetStatus(),
			NeedsReview: false, // Field not in proto TrackSummary
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetTrack(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	if idParam == "" {
		writeError(w, http.StatusBadRequest, "track id is required")
		return
	}

	track, err := s.db.ResolveTrack(&common.TrackId{ContentHash: idParam})
	if err != nil {
		writeError(w, http.StatusNotFound, "track not found")
		return
	}

	analysis, err := s.db.LatestCompleteAnalysis(track.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "analysis not found")
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

// ScanRequest is the JSON request for library scanning.
type ScanRequest struct {
	Roots       []string `json:"roots"`
	ForceRescan bool     `json:"force_rescan"`
}

// ScanResponse is the JSON response for library scanning.
type ScanResponse struct {
	Processed int64    `json:"processed"`
	Total     int64    `json:"total"`
	NewTracks []string `json:"new_tracks"`
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if len(req.Roots) == 0 {
		writeError(w, http.StatusBadRequest, "at least one root path is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	progress := make(chan scanner.ScanProgress)
	var scanErr error
	var newTrackIDs []int64
	var newPaths []string

	go func() {
		scanErr = s.scanner.Scan(ctx, req.Roots, req.ForceRescan, progress)
	}()

	var lastProcessed, lastTotal int64
	for p := range progress {
		if p.IsNew {
			newTrackIDs = append(newTrackIDs, p.TrackID)
			newPaths = append(newPaths, p.Path)
		}
		lastProcessed = p.Processed
		lastTotal = p.Total
	}

	if scanErr != nil && scanErr != context.Canceled {
		writeError(w, http.StatusInternalServerError, "scan failed: "+scanErr.Error())
		return
	}

	if len(newTrackIDs) > 0 {
		if err := s.scanner.EnqueueAnalysis(newTrackIDs, 0); err != nil {
			s.logger.Warn("failed to enqueue analysis jobs", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, ScanResponse{
		Processed: lastProcessed,
		Total:     lastTotal,
		NewTracks: newPaths,
	})
}

// AnalyzeRequest is the JSON request for track analysis.
type AnalyzeRequest struct {
	Paths    []string `json:"paths"`
	TrackIDs []string `json:"track_ids"`
	Force    bool     `json:"force"`
}

// AnalyzeResponse is the JSON response for track analysis.
type AnalyzeResponse struct {
	Analyzed []string `json:"analyzed"`
	Skipped  []string `json:"skipped"`
	Errors   []string `json:"errors"`
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if len(req.Paths) == 0 && len(req.TrackIDs) == 0 {
		writeError(w, http.StatusBadRequest, "paths or track_ids are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	var analyzed, skipped, errors []string

	// Process paths
	for _, path := range req.Paths {
		hash, err := scanner.ComputeHash(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		track, err := s.db.ResolveTrack(&common.TrackId{ContentHash: hash, Path: path})
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: track not found", path))
			continue
		}

		result, err := s.analyzeTrack(ctx, track, req.Force)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
		} else if result == "skipped" {
			skipped = append(skipped, path)
		} else {
			analyzed = append(analyzed, path)
		}
	}

	// Process track IDs
	for _, id := range req.TrackIDs {
		track, err := s.db.ResolveTrack(&common.TrackId{ContentHash: id})
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: track not found", id))
			continue
		}

		result, err := s.analyzeTrack(ctx, track, req.Force)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", id, err))
		} else if result == "skipped" {
			skipped = append(skipped, track.Path)
		} else {
			analyzed = append(analyzed, track.Path)
		}
	}

	writeJSON(w, http.StatusOK, AnalyzeResponse{
		Analyzed: analyzed,
		Skipped:  skipped,
		Errors:   errors,
	})
}

func (s *Server) analyzeTrack(ctx context.Context, track *storage.Track, force bool) (string, error) {
	version := int32(1)

	if !force {
		if rec, err := s.db.LatestAnalysisRecord(track.ID); err == nil && rec.Status == storage.AnalysisStatusComplete {
			return "skipped", nil
		}
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
		return "", err
	}

	rec, err := storage.AnalysisRecordFromProto(track.ID, version, res.GetAnalysis())
	if err != nil {
		return "", fmt.Errorf("marshal analysis failed: %w", err)
	}
	if err := s.db.UpsertAnalysis(rec); err != nil {
		return "", fmt.Errorf("persist analysis failed: %w", err)
	}

	return "analyzed", nil
}

// ProposeSetRequest is the JSON request for set planning.
type ProposeSetRequest struct {
	TrackIDs      []string `json:"track_ids"`
	Mode          string   `json:"mode"`
	AllowKeyJumps bool     `json:"allow_key_jumps"`
	MaxBpmStep    float64  `json:"max_bpm_step"`
	MustPlay      []string `json:"must_play"`
	Ban           []string `json:"ban"`
}

func (s *Server) handleProposeSet(w http.ResponseWriter, r *http.Request) {
	var req ProposeSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if len(req.TrackIDs) == 0 {
		writeError(w, http.StatusBadRequest, "track_ids are required")
		return
	}

	analyses := []*common.TrackAnalysis{}
	for _, id := range req.TrackIDs {
		track, err := s.db.ResolveTrack(&common.TrackId{ContentHash: id})
		if err != nil {
			writeError(w, http.StatusNotFound, fmt.Sprintf("track not found: %s", id))
			return
		}
		analysis, err := s.db.LatestCompleteAnalysis(track.ID)
		if err != nil {
			writeError(w, http.StatusPreconditionFailed, fmt.Sprintf("missing analysis for %s", track.Path))
			return
		}
		analyses = append(analyses, analysis)
	}

	mode := engine.SetMode_PEAK_TIME
	switch strings.ToUpper(req.Mode) {
	case "WARM_UP", "SET_MODE_WARM_UP":
		mode = engine.SetMode_WARM_UP
	case "OPEN_FORMAT", "SET_MODE_OPEN_FORMAT":
		mode = engine.SetMode_OPEN_FORMAT
	}

	mustPlay := make(map[string]bool)
	for _, h := range req.MustPlay {
		mustPlay[h] = true
	}
	ban := make(map[string]bool)
	for _, h := range req.Ban {
		ban[h] = true
	}

	opts := planner.Options{
		Mode:           mode,
		AllowKeyJumps:  req.AllowKeyJumps,
		MaxBpmStep:     req.MaxBpmStep,
		MustPlayHashes: mustPlay,
		BanHashes:      ban,
	}

	order, explanations, err := planner.Plan(analyses, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "set planning failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order":        order,
		"explanations": explanations,
	})
}

// ExportRequest is the JSON request for exporting a set.
type ExportRequest struct {
	TrackIDs     []string `json:"track_ids"`
	PlaylistName string   `json:"playlist_name"`
	OutputDir    string   `json:"output_dir"`
	Formats      []string `json:"formats"`
}

// ExportResponse is the JSON response for exporting a set.
type ExportResponse struct {
	PlaylistPath  string   `json:"playlist_path"`
	AnalysisJSON  string   `json:"analysis_json"`
	CuesCSV       string   `json:"cues_csv"`
	BundlePath    string   `json:"bundle_path"`
	VendorExports []string `json:"vendor_exports"`
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	var req ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if len(req.TrackIDs) == 0 {
		writeError(w, http.StatusBadRequest, "track_ids are required")
		return
	}

	outputDir := req.OutputDir
	if outputDir == "" {
		outputDir = fmt.Sprintf("%s/exports/%s", s.cfg.DataDir, time.Now().Format("20060102-150405"))
	}

	tracks := []exporter.TrackExport{}
	for _, id := range req.TrackIDs {
		track, err := s.db.ResolveTrack(&common.TrackId{ContentHash: id})
		if err != nil {
			writeError(w, http.StatusNotFound, fmt.Sprintf("track not found: %s", id))
			return
		}
		analysis, err := s.db.LatestCompleteAnalysis(track.ID)
		if err != nil {
			writeError(w, http.StatusPreconditionFailed, fmt.Sprintf("missing analysis for %s", track.Path))
			return
		}
		tracks = append(tracks, exporter.TrackExport{
			Path:     track.Path,
			Analysis: analysis,
		})
	}

	playlistName := req.PlaylistName
	if playlistName == "" {
		playlistName = "set"
	}

	result, err := exporter.WriteGeneric(outputDir, playlistName, tracks)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "export failed: "+err.Error())
		return
	}

	// Write vendor exports if requested
	vendorExports := []string{}
	for _, format := range req.Formats {
		switch strings.ToLower(format) {
		case "rekordbox":
			path, err := exporter.WriteRekordbox(outputDir, playlistName, tracks)
			if err != nil {
				s.logger.Warn("rekordbox export failed", "error", err)
			} else {
				vendorExports = append(vendorExports, path)
			}
		case "serato":
			path, err := exporter.WriteSerato(outputDir, playlistName, tracks)
			if err != nil {
				s.logger.Warn("serato export failed", "error", err)
			} else {
				vendorExports = append(vendorExports, path)
			}
		case "traktor":
			path, err := exporter.WriteTraktor(outputDir, playlistName, tracks)
			if err != nil {
				s.logger.Warn("traktor export failed", "error", err)
			} else {
				vendorExports = append(vendorExports, path)
			}
		}
	}

	writeJSON(w, http.StatusOK, ExportResponse{
		PlaylistPath:  result.PlaylistPath,
		AnalysisJSON:  result.AnalysisJSONPath,
		CuesCSV:       result.CuesCSVPath,
		BundlePath:    result.BundlePath,
		VendorExports: vendorExports,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
