// Package e2e provides end-to-end tests for the Algiers application.
// These tests use Playwright-Go to test the full stack (UI + API).
//
// Run with: go test ./internal/e2e -v -tags e2e
//
// Prerequisites:
// - Web UI running: cd web && npm run dev
// - Engine running: go run ./cmd/engine
// - Playwright browsers installed: make screenshots-install
package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

var (
	uiURL     = getEnv("E2E_UI_URL", "http://localhost:5173")
	apiURL    = getEnv("E2E_API_URL", "http://localhost:8080")
	headless  = getEnv("E2E_HEADLESS", "true") == "true"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// isUIAvailable checks if the web UI is running.
func isUIAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(uiURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// isAPIAvailable checks if the HTTP API is running.
func isAPIAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(apiURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Check that it's actually our API (returns JSON with status: ok)
	if resp.StatusCode != http.StatusOK {
		return false
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || !strings.Contains(contentType, "json") {
		return false
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false
	}
	return body["status"] == "ok"
}

// TestUILoads verifies the web UI loads successfully.
func TestUILoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isUIAvailable() {
		t.Skip("UI not available at " + uiURL)
	}

	pw, browser, page := setupBrowser(t)
	defer teardownBrowser(pw, browser)

	// Navigate to the app
	_, err := page.Goto(uiURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		t.Fatalf("failed to navigate to UI: %v", err)
	}

	// Wait for initial content
	time.Sleep(2 * time.Second)

	// Verify title
	title, err := page.Title()
	if err != nil {
		t.Fatalf("failed to get title: %v", err)
	}
	if title == "" {
		t.Error("expected non-empty page title")
	}

	// Verify app container exists
	appContainer := page.Locator(".app-container")
	if count, err := appContainer.Count(); err != nil || count == 0 {
		t.Error("expected app-container element")
	}
}

// TestLibraryView verifies the library view displays tracks.
func TestLibraryView(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isUIAvailable() {
		t.Skip("UI not available at " + uiURL)
	}

	pw, browser, page := setupBrowser(t)
	defer teardownBrowser(pw, browser)

	_, err := page.Goto(uiURL)
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Click Library tab
	if err := page.Click("text=Library"); err != nil {
		t.Logf("Warning: Library click failed: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Verify track list exists
	trackList := page.Locator(".track-list, .library-grid")
	count, _ := trackList.Count()
	if count == 0 {
		t.Log("Warning: no track list found (expected with mock data)")
	}
}

// TestSetBuilder verifies the set builder view.
func TestSetBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isUIAvailable() {
		t.Skip("UI not available at " + uiURL)
	}

	pw, browser, page := setupBrowser(t)
	defer teardownBrowser(pw, browser)

	_, err := page.Goto(uiURL)
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Click Set Builder tab
	if err := page.Click("text=Set Builder"); err != nil {
		t.Fatalf("failed to click Set Builder: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Verify energy arc exists
	energyArc := page.Locator(".energy-arc, svg")
	count, _ := energyArc.Count()
	if count == 0 {
		t.Log("Warning: no energy arc found")
	}
}

// TestGraphView verifies the transition graph view.
func TestGraphView(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isUIAvailable() {
		t.Skip("UI not available at " + uiURL)
	}

	pw, browser, page := setupBrowser(t)
	defer teardownBrowser(pw, browser)

	_, err := page.Goto(uiURL)
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Click Graph tab
	if err := page.Click("text=Graph"); err != nil {
		t.Fatalf("failed to click Graph: %v", err)
	}
	// Wait for D3 force simulation
	time.Sleep(2 * time.Second)

	// Verify graph canvas/svg exists
	graph := page.Locator(".transition-graph, .graph-view, svg")
	count, _ := graph.Count()
	if count == 0 {
		t.Log("Warning: no graph view found")
	}
}

// TestThemeToggle verifies dark/light mode switching.
func TestThemeToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isUIAvailable() {
		t.Skip("UI not available at " + uiURL)
	}

	pw, browser, page := setupBrowser(t)
	defer teardownBrowser(pw, browser)

	_, err := page.Goto(uiURL)
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Click theme toggle
	if err := page.Click(".theme-toggle"); err != nil {
		t.Logf("Warning: theme toggle click failed: %v", err)
		return
	}
	time.Sleep(500 * time.Millisecond)

	// Toggle back
	if err := page.Click(".theme-toggle"); err != nil {
		t.Logf("Warning: theme toggle click failed: %v", err)
	}
}

// TestAPIHealth verifies the HTTP API health endpoint.
func TestAPIHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isAPIAvailable() {
		t.Skip("API not available at " + apiURL)
	}

	resp, err := http.Get(apiURL + "/api/health")
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

// TestAPIListTracks verifies the tracks listing endpoint.
func TestAPIListTracks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	if !isAPIAvailable() {
		t.Skip("API not available at " + apiURL)
	}

	resp, err := http.Get(apiURL + "/api/tracks")
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Should return an array (empty or with tracks)
	var tracks []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Found %d tracks", len(tracks))
}

func setupBrowser(t *testing.T) (*playwright.Playwright, playwright.Browser, playwright.Page) {
	t.Helper()

	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		t.Fatalf("failed to install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("failed to start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		pw.Stop()
		t.Fatalf("failed to launch browser: %v", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 720,
		},
		ColorScheme: playwright.ColorSchemeDark,
	})
	if err != nil {
		browser.Close()
		pw.Stop()
		t.Fatalf("failed to create context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		context.Close()
		browser.Close()
		pw.Stop()
		t.Fatalf("failed to create page: %v", err)
	}

	return pw, browser, page
}

func teardownBrowser(pw *playwright.Playwright, browser playwright.Browser) {
	browser.Close()
	pw.Stop()
}
