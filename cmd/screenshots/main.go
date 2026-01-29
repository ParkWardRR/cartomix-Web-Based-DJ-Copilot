// Command screenshots captures UI screenshots using Playwright-Go.
// This is used for README assets and visual regression testing.
//
// Usage:
//
//	go run ./cmd/screenshots [flags]
//
// Flags:
//
//	-url        Base URL of the web UI (default: http://localhost:5173)
//	-out        Output directory for screenshots (default: docs/assets/screens)
//	-headless   Run browser in headless mode (default: true)
//	-width      Viewport width (default: 1920)
//	-height     Viewport height (default: 1080)
//	-scale      Device scale factor for retina (default: 2)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Config struct {
	URL      string
	OutDir   string
	Headless bool
	Width    int
	Height   int
	Scale    float64
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.URL, "url", "http://localhost:5173", "Base URL of the web UI")
	flag.StringVar(&cfg.OutDir, "out", "docs/assets/screens", "Output directory for screenshots")
	flag.BoolVar(&cfg.Headless, "headless", true, "Run browser in headless mode")
	flag.IntVar(&cfg.Width, "width", 1920, "Viewport width")
	flag.IntVar(&cfg.Height, "height", 1080, "Viewport height")
	flag.Float64Var(&cfg.Scale, "scale", 2, "Device scale factor (2 for retina)")
	flag.Parse()

	if err := run(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(cfg Config) error {
	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// Install Playwright browsers if needed
	if err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	}); err != nil {
		return fmt.Errorf("install playwright: %w", err)
	}

	// Launch Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("start playwright: %w", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(cfg.Headless),
	})
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	defer browser.Close()

	// Create context with viewport settings
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  cfg.Width,
			Height: cfg.Height,
		},
		DeviceScaleFactor: playwright.Float(cfg.Scale),
		ColorScheme:       playwright.ColorSchemeDark,
	})
	if err != nil {
		return fmt.Errorf("create context: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	// Define screenshots to capture
	screenshots := []struct {
		name        string
		setup       func(page playwright.Page) error
		description string
	}{
		{
			name:        "algiers-library-view.png",
			description: "Library View with track grid and filters",
			setup: func(p playwright.Page) error {
				// Default view, just wait for load
				time.Sleep(2 * time.Second)
				return nil
			},
		},
		{
			name:        "algiers-hero.png",
			description: "Active waveform with spectrum analyzer",
			setup: func(p playwright.Page) error {
				// Click play button to activate spectrum
				playBtn := p.Locator(".transport-btn.play")
				visible, _ := playBtn.IsVisible()
				if visible {
					if err := playBtn.Click(); err != nil {
						return err
					}
					time.Sleep(1 * time.Second)
				}
				return nil
			},
		},
		{
			name:        "algiers-set-builder.png",
			description: "Set Builder with energy arc visualization",
			setup: func(p playwright.Page) error {
				if err := p.Click("text=Set Builder"); err != nil {
					return err
				}
				time.Sleep(1500 * time.Millisecond)
				return nil
			},
		},
		{
			name:        "algiers-graph-view.png",
			description: "Transition Graph (D3.js force-directed)",
			setup: func(p playwright.Page) error {
				if err := p.Click("text=Graph View"); err != nil {
					return err
				}
				// Wait for D3 force simulation to settle
				time.Sleep(2500 * time.Millisecond)
				return nil
			},
		},
		{
			name:        "algiers-light-mode.png",
			description: "Light mode theme",
			setup: func(p playwright.Page) error {
				if err := p.Click(".theme-toggle"); err != nil {
					return err
				}
				time.Sleep(500 * time.Millisecond)
				return nil
			},
		},
	}

	// Navigate to the app
	log.Printf("Navigating to %s", cfg.URL)
	if _, err := page.Goto(cfg.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return fmt.Errorf("navigate to %s: %w", cfg.URL, err)
	}

	// Wait for initial load
	time.Sleep(2 * time.Second)

	// Capture each screenshot
	for _, shot := range screenshots {
		log.Printf("Capturing: %s - %s", shot.name, shot.description)

		if err := shot.setup(page); err != nil {
			log.Printf("Warning: setup for %s failed: %v", shot.name, err)
		}

		outPath := filepath.Join(cfg.OutDir, shot.name)
		if _, err := page.Screenshot(playwright.PageScreenshotOptions{
			Path:     playwright.String(outPath),
			FullPage: playwright.Bool(false),
		}); err != nil {
			return fmt.Errorf("screenshot %s: %w", shot.name, err)
		}

		log.Printf("  Saved: %s", outPath)
	}

	log.Println("All screenshots captured successfully!")
	return nil
}
