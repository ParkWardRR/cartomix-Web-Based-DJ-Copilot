// Command screenshots captures UI screenshots and GIF animations using Playwright-Go.
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
//	-width      Viewport width (default: 1280)
//	-height     Viewport height (default: 720)
//	-scale      Device scale factor for retina (default: 2)
//	-gif        Capture animated GIF hero (default: true)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	CaptureGIF bool
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.URL, "url", "http://localhost:5173", "Base URL of the web UI")
	flag.StringVar(&cfg.OutDir, "out", "docs/assets/screens", "Output directory for screenshots")
	flag.BoolVar(&cfg.Headless, "headless", true, "Run browser in headless mode")
	flag.IntVar(&cfg.Width, "width", 1280, "Viewport width")
	flag.IntVar(&cfg.Height, "height", 720, "Viewport height")
	flag.Float64Var(&cfg.Scale, "scale", 2, "Device scale factor (2 for retina)")
	flag.BoolVar(&cfg.CaptureGIF, "gif", true, "Capture animated GIF hero")
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

	// Create context with viewport and video recording for GIF
	videoDir := filepath.Join(cfg.OutDir, "video")
	if cfg.CaptureGIF {
		os.MkdirAll(videoDir, 0755)
	}

	contextOpts := playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  cfg.Width,
			Height: cfg.Height,
		},
		DeviceScaleFactor: playwright.Float(cfg.Scale),
		ColorScheme:       playwright.ColorSchemeDark,
	}

	if cfg.CaptureGIF {
		contextOpts.RecordVideo = &playwright.RecordVideo{
			Dir: videoDir,
			Size: &playwright.Size{
				Width:  cfg.Width,
				Height: cfg.Height,
			},
		}
	}

	context, err := browser.NewContext(contextOpts)
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
				// Click play to activate spectrum
				playBtn := p.Locator(".transport-btn.play")
				if err := playBtn.Click(); err != nil {
					log.Printf("Play button click: %v", err)
				}
				time.Sleep(1500 * time.Millisecond)
				return nil
			},
		},
		{
			name:        "algiers-hero.png",
			description: "Active waveform with spectrum analyzer",
			setup: func(p playwright.Page) error {
				// Already playing from previous, wait for animation
				time.Sleep(500 * time.Millisecond)
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
				time.Sleep(1000 * time.Millisecond)
				return nil
			},
		},
		{
			name:        "algiers-graph-view.png",
			description: "Transition Graph (D3.js force-directed)",
			setup: func(p playwright.Page) error {
				if err := p.Click("text=Graph"); err != nil {
					return err
				}
				// Wait for D3 force simulation to settle
				time.Sleep(2000 * time.Millisecond)
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

	// Capture GIF animation
	if cfg.CaptureGIF {
		log.Println("Capturing GIF animation...")

		// Go back to dark mode library view
		if err := page.Click(".theme-toggle"); err != nil {
			log.Printf("Theme toggle: %v", err)
		}
		time.Sleep(500 * time.Millisecond)

		if err := page.Click("text=Library"); err != nil {
			log.Printf("Library click: %v", err)
		}
		time.Sleep(1000 * time.Millisecond)

		// Record the animation sequence
		// Click play and let it animate
		playBtn := page.Locator(".transport-btn.play")
		if err := playBtn.Click(); err != nil {
			log.Printf("Play click: %v", err)
		}

		// Let it record for a few seconds
		time.Sleep(4 * time.Second)

		// Stop recording by closing the page
		page.Close()

		// Get the video file
		video := page.Video()
		if video != nil {
			videoPath, err := video.Path()
			if err == nil {
				log.Printf("Video recorded: %s", videoPath)

				// Convert to GIF using ffmpeg
				gifPath := filepath.Join(cfg.OutDir, "algiers-demo.gif")
				if err := convertToGIF(videoPath, gifPath, cfg.Width/2); err != nil {
					log.Printf("Warning: GIF conversion failed: %v", err)
					log.Println("To convert manually: ffmpeg -i video.webm -vf \"fps=12,scale=640:-1:flags=lanczos\" -loop 0 demo.gif")
				} else {
					log.Printf("GIF saved: %s", gifPath)
				}
			}
		}
	}

	log.Println("All screenshots captured successfully!")
	return nil
}

func convertToGIF(videoPath, gifPath string, width int) error {
	// Use ffmpeg to convert video to GIF
	// -vf: video filters - fps, scale, palettegen/paletteuse for quality
	cmd := exec.Command("ffmpeg",
		"-y", // overwrite output
		"-i", videoPath,
		"-vf", fmt.Sprintf("fps=12,scale=%d:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse", width),
		"-loop", "0",
		gifPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
