package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/cartomix/cancun/internal/similarity"
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
	return deprecationMiddleware(corsMiddleware(s.mux))
}

// deprecationMiddleware adds headers indicating the HTTP API is deprecated in favor of gRPC.
func deprecationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// RFC 8594: Sunset Header
		w.Header().Set("Sunset", "Wed, 01 Jul 2026 00:00:00 GMT")
		w.Header().Set("Deprecation", "true")
		w.Header().Set("X-API-Deprecation-Notice", "This HTTP REST API is deprecated. Please migrate to gRPC for improved performance. See docs/gRPC-MIGRATION.md")
		w.Header().Set("Link", `</docs/gRPC-MIGRATION.md>; rel="deprecation"; type="text/markdown"`)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/tracks", s.handleListTracks)
	s.mux.HandleFunc("GET /api/tracks/{id}", s.handleGetTrack)
	s.mux.HandleFunc("GET /api/tracks/{id}/similar", s.handleSimilarTracks)
	s.mux.HandleFunc("POST /api/scan", s.handleScan)
	s.mux.HandleFunc("POST /api/analyze", s.handleAnalyze)
	s.mux.HandleFunc("POST /api/set/propose", s.handleProposeSet)
	s.mux.HandleFunc("POST /api/export", s.handleExport)
	s.mux.HandleFunc("GET /api/ml/settings", s.handleGetMLSettings)
	s.mux.HandleFunc("PUT /api/ml/settings", s.handleUpdateMLSettings)

	// Training endpoints
	s.mux.HandleFunc("GET /api/training/labels", s.handleListTrainingLabels)
	s.mux.HandleFunc("POST /api/training/labels", s.handleAddTrainingLabel)
	s.mux.HandleFunc("DELETE /api/training/labels/{id}", s.handleDeleteTrainingLabel)
	s.mux.HandleFunc("GET /api/training/labels/stats", s.handleTrainingLabelStats)
	s.mux.HandleFunc("POST /api/training/start", s.handleStartTraining)
	s.mux.HandleFunc("GET /api/training/jobs", s.handleListTrainingJobs)
	s.mux.HandleFunc("GET /api/training/jobs/{id}", s.handleGetTrainingJob)
	s.mux.HandleFunc("GET /api/training/models", s.handleListModelVersions)
	s.mux.HandleFunc("POST /api/training/models/{version}/activate", s.handleActivateModel)
	s.mux.HandleFunc("DELETE /api/training/models/{version}", s.handleDeleteModel)

	// Audio streaming endpoint
	s.mux.HandleFunc("GET /api/audio", s.handleAudio)

	// Static file serving (for standalone app mode)
	if s.cfg.WebRoot != "" {
		s.mux.HandleFunc("/", s.handleStaticFiles)
	}
}

// handleStaticFiles serves static files from WebRoot with SPA fallback to index.html
func (s *Server) handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Don't serve static files for API routes
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.cfg.WebRoot, r.URL.Path)

	// Check if file exists
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		http.ServeFile(w, r, path)
		return
	}

	// SPA fallback: serve index.html for non-file routes
	indexPath := filepath.Join(s.cfg.WebRoot, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	http.NotFound(w, r)
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

// SimilarTracksResponse is the JSON response for similar tracks.
type SimilarTracksResponse struct {
	Query   TrackSummaryResponse        `json:"query"`
	Similar []similarity.SimilarityResult `json:"similar"`
}

func (s *Server) handleSimilarTracks(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	if idParam == "" {
		writeError(w, http.StatusBadRequest, "track id is required")
		return
	}

	// Parse limit from query string (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Resolve query track
	track, err := s.db.ResolveTrack(&common.TrackId{ContentHash: idParam})
	if err != nil {
		writeError(w, http.StatusNotFound, "track not found")
		return
	}

	// Get query track features
	queryFeatures, err := s.db.GetTrackFeaturesForSimilarity(track.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "track analysis not found")
		return
	}

	// Check if track has OpenL3 embedding
	if len(queryFeatures.OpenL3Embedding) == 0 {
		writeError(w, http.StatusPreconditionFailed, "track has no ML embedding - re-analyze with OpenL3 enabled")
		return
	}

	// Get all candidate tracks (excluding query)
	candidates, err := s.db.GetTrackFeaturesExcluding([]int64{track.ID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch candidates: "+err.Error())
		return
	}

	if len(candidates) == 0 {
		writeJSON(w, http.StatusOK, SimilarTracksResponse{
			Query: TrackSummaryResponse{
				ContentHash: track.ContentHash,
				Path:        track.Path,
				Title:       queryFeatures.Title,
				Artist:      queryFeatures.Artist,
				BPM:         queryFeatures.BPM,
				Key:         queryFeatures.KeyValue,
				Energy:      queryFeatures.Energy,
			},
			Similar: []similarity.SimilarityResult{},
		})
		return
	}

	// Find similar tracks
	similar := similarity.FindSimilar(queryFeatures, candidates, limit)

	// Cache results for future queries
	for _, sim := range similar {
		_ = s.db.CacheSimilarity(
			track.ID, sim.TrackID,
			sim.VibeMatch/100, sim.Score,
			sim.TempoMatch/100, sim.KeyMatch/100, sim.EnergyMatch/100,
			sim.Explanation,
		)
	}

	writeJSON(w, http.StatusOK, SimilarTracksResponse{
		Query: TrackSummaryResponse{
			ContentHash: track.ContentHash,
			Path:        track.Path,
			Title:       queryFeatures.Title,
			Artist:      queryFeatures.Artist,
			BPM:         queryFeatures.BPM,
			Key:         queryFeatures.KeyValue,
			Energy:      queryFeatures.Energy,
		},
		Similar: similar,
	})
}

// MLSettingsResponse is the JSON response for ML settings.
type MLSettingsResponse struct {
	OpenL3Enabled         bool    `json:"openl3_enabled"`
	SoundAnalysisEnabled  bool    `json:"sound_analysis_enabled"`
	CustomModelEnabled    bool    `json:"custom_model_enabled"`
	MinSimilarityThreshold float64 `json:"min_similarity_threshold"`
	ShowExplanations      bool    `json:"show_explanations"`
}

func (s *Server) handleGetMLSettings(w http.ResponseWriter, _ *http.Request) {
	settings, err := s.db.GetMLSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings: "+err.Error())
		return
	}

	response := MLSettingsResponse{
		OpenL3Enabled:         settings["openl3_enabled"] == "true",
		SoundAnalysisEnabled:  settings["sound_analysis_enabled"] == "true",
		CustomModelEnabled:    settings["custom_model_enabled"] == "true",
		MinSimilarityThreshold: 0.5,
		ShowExplanations:      settings["show_explanations"] != "false",
	}

	if threshold, ok := settings["min_similarity_threshold"]; ok {
		if val, err := strconv.ParseFloat(threshold, 64); err == nil {
			response.MinSimilarityThreshold = val
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// MLSettingsRequest is the JSON request for updating ML settings.
type MLSettingsRequest struct {
	OpenL3Enabled         *bool    `json:"openl3_enabled,omitempty"`
	SoundAnalysisEnabled  *bool    `json:"sound_analysis_enabled,omitempty"`
	CustomModelEnabled    *bool    `json:"custom_model_enabled,omitempty"`
	MinSimilarityThreshold *float64 `json:"min_similarity_threshold,omitempty"`
	ShowExplanations      *bool    `json:"show_explanations,omitempty"`
}

func (s *Server) handleUpdateMLSettings(w http.ResponseWriter, r *http.Request) {
	var req MLSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.OpenL3Enabled != nil {
		_ = s.db.SetMLSetting("openl3_enabled", fmt.Sprintf("%t", *req.OpenL3Enabled))
	}
	if req.SoundAnalysisEnabled != nil {
		_ = s.db.SetMLSetting("sound_analysis_enabled", fmt.Sprintf("%t", *req.SoundAnalysisEnabled))
	}
	if req.CustomModelEnabled != nil {
		_ = s.db.SetMLSetting("custom_model_enabled", fmt.Sprintf("%t", *req.CustomModelEnabled))
	}
	if req.MinSimilarityThreshold != nil {
		_ = s.db.SetMLSetting("min_similarity_threshold", fmt.Sprintf("%.2f", *req.MinSimilarityThreshold))
	}
	if req.ShowExplanations != nil {
		_ = s.db.SetMLSetting("show_explanations", fmt.Sprintf("%t", *req.ShowExplanations))
	}

	// Return updated settings
	s.handleGetMLSettings(w, r)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// Training API Types

// TrainingLabelRequest is the JSON request for adding a training label.
type TrainingLabelRequest struct {
	TrackID          int64   `json:"track_id"`
	LabelValue       string  `json:"label_value"`
	StartBeat        int     `json:"start_beat"`
	EndBeat          int     `json:"end_beat"`
	StartTimeSeconds float64 `json:"start_time_seconds"`
	EndTimeSeconds   float64 `json:"end_time_seconds"`
	Source           string  `json:"source,omitempty"`
}

// TrainingLabelResponse is the JSON response for training labels.
type TrainingLabelResponse struct {
	ID               int64   `json:"id"`
	TrackID          int64   `json:"track_id"`
	ContentHash      string  `json:"content_hash"`
	TrackPath        string  `json:"track_path"`
	LabelValue       string  `json:"label_value"`
	StartBeat        int     `json:"start_beat"`
	EndBeat          int     `json:"end_beat"`
	StartTimeSeconds float64 `json:"start_time_seconds"`
	EndTimeSeconds   float64 `json:"end_time_seconds"`
	Source           string  `json:"source"`
	CreatedAt        string  `json:"created_at"`
}

// TrainingJobResponse is the JSON response for training jobs.
type TrainingJobResponse struct {
	JobID        string         `json:"job_id"`
	Status       string         `json:"status"`
	Progress     float64        `json:"progress"`
	CurrentEpoch *int           `json:"current_epoch,omitempty"`
	TotalEpochs  *int           `json:"total_epochs,omitempty"`
	CurrentLoss  *float64       `json:"current_loss,omitempty"`
	Accuracy     *float64       `json:"accuracy,omitempty"`
	F1Score      *float64       `json:"f1_score,omitempty"`
	ModelPath    *string        `json:"model_path,omitempty"`
	ModelVersion *int           `json:"model_version,omitempty"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	LabelCounts  map[string]int `json:"label_counts,omitempty"`
	StartedAt    *string        `json:"started_at,omitempty"`
	CompletedAt  *string        `json:"completed_at,omitempty"`
	CreatedAt    string         `json:"created_at"`
}

// ModelVersionResponse is the JSON response for model versions.
type ModelVersionResponse struct {
	ID            int64          `json:"id"`
	ModelType     string         `json:"model_type"`
	Version       int            `json:"version"`
	ModelPath     string         `json:"model_path"`
	Accuracy      float64        `json:"accuracy"`
	F1Score       float64        `json:"f1_score"`
	IsActive      bool           `json:"is_active"`
	LabelCounts   map[string]int `json:"label_counts,omitempty"`
	TrainingJobID *string        `json:"training_job_id,omitempty"`
	CreatedAt     string         `json:"created_at"`
}

// TrainingLabelStatsResponse is the JSON response for training label statistics.
type TrainingLabelStatsResponse struct {
	TotalLabels        int            `json:"total_labels"`
	LabelCounts        map[string]int `json:"label_counts"`
	TracksCovered      int            `json:"tracks_covered"`
	AvgPerTrack        float64        `json:"avg_per_track"`
	ReadyForTraining   bool           `json:"ready_for_training"`
	MinSamplesRequired int            `json:"min_samples_required"`
}

func (s *Server) handleListTrainingLabels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse optional filters
	var trackID *int64
	var labelValue *string

	if tidStr := r.URL.Query().Get("track_id"); tidStr != "" {
		if tid, err := strconv.ParseInt(tidStr, 10, 64); err == nil {
			trackID = &tid
		}
	}
	if lv := r.URL.Query().Get("label"); lv != "" {
		labelValue = &lv
	}

	labels, err := s.db.GetTrainingLabels(ctx, trackID, labelValue)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get training labels: "+err.Error())
		return
	}

	response := make([]TrainingLabelResponse, 0, len(labels))
	for _, l := range labels {
		response = append(response, TrainingLabelResponse{
			ID:               l.ID,
			TrackID:          l.TrackID,
			ContentHash:      l.ContentHash,
			TrackPath:        l.TrackPath,
			LabelValue:       l.LabelValue,
			StartBeat:        l.StartBeat,
			EndBeat:          l.EndBeat,
			StartTimeSeconds: l.StartTimeSeconds,
			EndTimeSeconds:   l.EndTimeSeconds,
			Source:           l.Source,
			CreatedAt:        l.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleAddTrainingLabel(w http.ResponseWriter, r *http.Request) {
	var req TrainingLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Validate label value
	validLabels := map[string]bool{
		"intro": true, "build": true, "drop": true, "break": true,
		"outro": true, "verse": true, "chorus": true,
	}
	if !validLabels[req.LabelValue] {
		writeError(w, http.StatusBadRequest, "invalid label_value: must be one of intro, build, drop, break, outro, verse, chorus")
		return
	}

	source := req.Source
	if source == "" {
		source = "user"
	}

	label := &storage.TrainingLabel{
		TrackID:          req.TrackID,
		LabelValue:       req.LabelValue,
		StartBeat:        req.StartBeat,
		EndBeat:          req.EndBeat,
		StartTimeSeconds: req.StartTimeSeconds,
		EndTimeSeconds:   req.EndTimeSeconds,
		Source:           source,
	}

	if err := s.db.AddTrainingLabel(r.Context(), label); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add training label: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      label.ID,
		"message": "label added successfully",
	})
}

func (s *Server) handleDeleteTrainingLabel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid label id")
		return
	}

	if err := s.db.DeleteTrainingLabel(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete label: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "label deleted"})
}

func (s *Server) handleTrainingLabelStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetTrainingLabelStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats: "+err.Error())
		return
	}

	minSamples := 10 // Minimum samples per class
	readyForTraining := true
	for _, count := range stats.LabelCounts {
		if count < minSamples {
			readyForTraining = false
			break
		}
	}

	// Need at least 2 classes
	if len(stats.LabelCounts) < 2 {
		readyForTraining = false
	}

	writeJSON(w, http.StatusOK, TrainingLabelStatsResponse{
		TotalLabels:        stats.TotalLabels,
		LabelCounts:        stats.LabelCounts,
		TracksCovered:      stats.TracksCovered,
		AvgPerTrack:        stats.AvgPerTrack,
		ReadyForTraining:   readyForTraining,
		MinSamplesRequired: minSamples,
	})
}

func (s *Server) handleStartTraining(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get current label stats
	stats, err := s.db.GetTrainingLabelStats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get training data: "+err.Error())
		return
	}

	// Validate we have enough data
	minSamples := 10
	for label, count := range stats.LabelCounts {
		if count < minSamples {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("need at least %d samples for %s (have %d)", minSamples, label, count))
			return
		}
	}

	if len(stats.LabelCounts) < 2 {
		writeError(w, http.StatusBadRequest, "need at least 2 different label types")
		return
	}

	// Create training job
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())
	if err := s.db.CreateTrainingJob(ctx, jobID, stats.LabelCounts); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create training job: "+err.Error())
		return
	}

	// Start training in background (would call Swift analyzer in production)
	go s.runTrainingJob(jobID)

	writeJSON(w, http.StatusAccepted, map[string]string{
		"job_id":  jobID,
		"message": "training job started",
	})
}

func (s *Server) runTrainingJob(jobID string) {
	ctx := context.Background()

	// Update progress
	s.db.UpdateTrainingJobProgress(ctx, jobID, "preparing", 0.1, nil, nil, nil)

	// Simulate training stages (in production, this would call the Swift trainer)
	stages := []struct {
		status   string
		progress float64
		delay    time.Duration
	}{
		{"preparing", 0.2, 500 * time.Millisecond},
		{"training", 0.4, 1 * time.Second},
		{"training", 0.6, 1 * time.Second},
		{"training", 0.8, 1 * time.Second},
		{"evaluating", 0.9, 500 * time.Millisecond},
	}

	for _, stage := range stages {
		time.Sleep(stage.delay)
		s.db.UpdateTrainingJobProgress(ctx, jobID, stage.status, stage.progress, nil, nil, nil)
	}

	// Get next model version
	versions, _ := s.db.GetModelVersions(ctx, "dj_section")
	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[0].Version + 1
	}

	// Complete training with mock results
	accuracy := 0.85 + float64(nextVersion)*0.01 // Improve slightly with each version
	if accuracy > 0.95 {
		accuracy = 0.95
	}
	f1Score := accuracy - 0.02
	modelPath := fmt.Sprintf("/models/dj_section_v%d.mlmodelc", nextVersion)

	s.db.CompleteTrainingJob(ctx, jobID, accuracy, f1Score, modelPath, nextVersion)

	// Add model version
	stats, _ := s.db.GetTrainingLabelStats(ctx)
	s.db.AddModelVersion(ctx, &storage.ModelVersion{
		ModelType:     "dj_section",
		Version:       nextVersion,
		ModelPath:     modelPath,
		Accuracy:      accuracy,
		F1Score:       f1Score,
		IsActive:      false,
		LabelCounts:   stats.LabelCounts,
		TrainingJobID: &jobID,
	})

	s.logger.Info("training completed", "job_id", jobID, "version", nextVersion, "accuracy", accuracy)
}

func (s *Server) handleListTrainingJobs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	jobs, err := s.db.ListTrainingJobs(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list jobs: "+err.Error())
		return
	}

	response := make([]TrainingJobResponse, 0, len(jobs))
	for _, j := range jobs {
		resp := TrainingJobResponse{
			JobID:        j.JobID,
			Status:       j.Status,
			Progress:     j.Progress,
			CurrentEpoch: j.CurrentEpoch,
			TotalEpochs:  j.TotalEpochs,
			CurrentLoss:  j.CurrentLoss,
			Accuracy:     j.Accuracy,
			F1Score:      j.F1Score,
			ModelPath:    j.ModelPath,
			ModelVersion: j.ModelVersion,
			ErrorMessage: j.ErrorMessage,
			LabelCounts:  j.LabelCounts,
			CreatedAt:    j.CreatedAt.Format(time.RFC3339),
		}
		if j.StartedAt != nil {
			s := j.StartedAt.Format(time.RFC3339)
			resp.StartedAt = &s
		}
		if j.CompletedAt != nil {
			s := j.CompletedAt.Format(time.RFC3339)
			resp.CompletedAt = &s
		}
		response = append(response, resp)
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetTrainingJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "missing job id")
		return
	}

	job, err := s.db.GetTrainingJob(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get job: "+err.Error())
		return
	}
	if job == nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	resp := TrainingJobResponse{
		JobID:        job.JobID,
		Status:       job.Status,
		Progress:     job.Progress,
		CurrentEpoch: job.CurrentEpoch,
		TotalEpochs:  job.TotalEpochs,
		CurrentLoss:  job.CurrentLoss,
		Accuracy:     job.Accuracy,
		F1Score:      job.F1Score,
		ModelPath:    job.ModelPath,
		ModelVersion: job.ModelVersion,
		ErrorMessage: job.ErrorMessage,
		LabelCounts:  job.LabelCounts,
		CreatedAt:    job.CreatedAt.Format(time.RFC3339),
	}
	if job.StartedAt != nil {
		s := job.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if job.CompletedAt != nil {
		s := job.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &s
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListModelVersions(w http.ResponseWriter, r *http.Request) {
	modelType := r.URL.Query().Get("type")
	if modelType == "" {
		modelType = "dj_section"
	}

	versions, err := s.db.GetModelVersions(r.Context(), modelType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list models: "+err.Error())
		return
	}

	response := make([]ModelVersionResponse, 0, len(versions))
	for _, v := range versions {
		response = append(response, ModelVersionResponse{
			ID:            v.ID,
			ModelType:     v.ModelType,
			Version:       v.Version,
			ModelPath:     v.ModelPath,
			Accuracy:      v.Accuracy,
			F1Score:       v.F1Score,
			IsActive:      v.IsActive,
			LabelCounts:   v.LabelCounts,
			TrainingJobID: v.TrainingJobID,
			CreatedAt:     v.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleActivateModel(w http.ResponseWriter, r *http.Request) {
	versionStr := r.PathValue("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid version")
		return
	}

	if err := s.db.ActivateModelVersion(r.Context(), "dj_section", version); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to activate model: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "model activated",
		"version": version,
	})
}

func (s *Server) handleDeleteModel(w http.ResponseWriter, r *http.Request) {
	versionStr := r.PathValue("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid version")
		return
	}

	// Don't allow deleting active model
	active, _ := s.db.GetActiveModelVersion(r.Context(), "dj_section")
	if active != nil && active.Version == version {
		writeError(w, http.StatusBadRequest, "cannot delete active model")
		return
	}

	if err := s.db.DeleteModelVersion(r.Context(), "dj_section", version); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete model: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "model deleted"})
}

// handleAudio streams audio files for playback in the UI.
// It supports HTTP Range requests for seeking.
func (s *Server) handleAudio(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path parameter is required")
		return
	}

	// Verify the path exists and is a file
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "audio file not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to access file: "+err.Error())
		}
		return
	}

	if info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is a directory, not a file")
		return
	}

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	contentType := "application/octet-stream"
	switch ext {
	case ".mp3":
		contentType = "audio/mpeg"
	case ".wav":
		contentType = "audio/wav"
	case ".flac":
		contentType = "audio/flac"
	case ".aac", ".m4a":
		contentType = "audio/aac"
	case ".ogg":
		contentType = "audio/ogg"
	case ".aiff", ".aif":
		contentType = "audio/aiff"
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to open file: "+err.Error())
		return
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, max-age=86400")

	// Use http.ServeContent for Range request support
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
}
