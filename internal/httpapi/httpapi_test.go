package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a minimal server for testing the health endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestCORSMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := corsMiddleware(inner)

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header to allow all origins")
	}

	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestTrackSummaryResponseJSON(t *testing.T) {
	response := TrackSummaryResponse{
		ContentHash: "abc123",
		Path:        "/music/track.mp3",
		Title:       "Test Track",
		Artist:      "Test Artist",
		BPM:         128.5,
		Key:         "8A",
		Energy:      7,
		CueCount:    4,
		Status:      "analyzed",
		NeedsReview: false,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded TrackSummaryResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ContentHash != response.ContentHash {
		t.Errorf("content_hash mismatch: got %s, want %s", decoded.ContentHash, response.ContentHash)
	}
	if decoded.BPM != response.BPM {
		t.Errorf("bpm mismatch: got %f, want %f", decoded.BPM, response.BPM)
	}
	if decoded.Key != response.Key {
		t.Errorf("key mismatch: got %s, want %s", decoded.Key, response.Key)
	}
}

func TestScanRequestJSON(t *testing.T) {
	request := ScanRequest{
		Roots:       []string{"/music/folder1", "/music/folder2"},
		ForceRescan: true,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ScanRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Roots) != 2 {
		t.Errorf("expected 2 roots, got %d", len(decoded.Roots))
	}
	if !decoded.ForceRescan {
		t.Error("expected force_rescan to be true")
	}
}

func TestProposeSetRequestJSON(t *testing.T) {
	request := ProposeSetRequest{
		TrackIDs:      []string{"hash1", "hash2", "hash3"},
		Mode:          "WARM_UP",
		AllowKeyJumps: false,
		MaxBpmStep:    8.0,
		MustPlay:      []string{"hash1"},
		Ban:           []string{},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ProposeSetRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.TrackIDs) != 3 {
		t.Errorf("expected 3 track_ids, got %d", len(decoded.TrackIDs))
	}
	if decoded.Mode != "WARM_UP" {
		t.Errorf("expected mode WARM_UP, got %s", decoded.Mode)
	}
	if decoded.MaxBpmStep != 8.0 {
		t.Errorf("expected max_bpm_step 8.0, got %f", decoded.MaxBpmStep)
	}
}

func TestExportRequestJSON(t *testing.T) {
	request := ExportRequest{
		TrackIDs:     []string{"hash1", "hash2"},
		PlaylistName: "my-set",
		OutputDir:    "/exports/2026-01",
		Formats:      []string{"rekordbox", "serato", "traktor"},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ExportRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.PlaylistName != "my-set" {
		t.Errorf("expected playlist_name 'my-set', got %s", decoded.PlaylistName)
	}
	if len(decoded.Formats) != 3 {
		t.Errorf("expected 3 formats, got %d", len(decoded.Formats))
	}
}
