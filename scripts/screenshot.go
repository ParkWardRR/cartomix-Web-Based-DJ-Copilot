package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	if err := playwright.Install(); err != nil {
		log.Fatalf("install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		log.Fatalf("launch browser: %v", err)
	}
	defer browser.Close()

	// Ensure docs/assets/screens exists
	os.MkdirAll("docs/assets/screens", 0755)
	os.MkdirAll("docs/assets/screens/video", 0755)

	// Create context with video recording enabled
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 1920, Height: 1080},
		RecordVideo: &playwright.RecordVideo{
			Dir:  "docs/assets/screens/video",
			Size: &playwright.Size{Width: 1920, Height: 1080},
		},
	})
	if err != nil {
		log.Fatalf("new context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("new page: %v", err)
	}

	// Wait for app to fully load
	log.Println("Loading app...")
	if _, err = page.Goto("http://localhost:5173", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("goto: %v", err)
	}

	// Wait for React to render with demo data
	time.Sleep(2 * time.Second)

	log.Println("🎬 Recording demo video...")

	// === DEMO SEQUENCE ===

	// 1. Show Library view - click through some tracks
	log.Println("  📚 Library view...")
	time.Sleep(1500 * time.Millisecond)

	// Click on different tracks to show selection
	page.Click("text=Pulse Drive")
	time.Sleep(800 * time.Millisecond)
	page.Click("text=Neon Rush")
	time.Sleep(800 * time.Millisecond)
	page.Click("text=Berghain Sunrise")
	time.Sleep(1000 * time.Millisecond)

	// 2. Switch to Set Builder
	log.Println("  🎛️ Set Builder view...")
	page.Click("text=Set Builder")
	time.Sleep(2000 * time.Millisecond)

	// Scroll through the set list
	page.Click("text=Enrico Sangiuliano")
	time.Sleep(800 * time.Millisecond)
	page.Click("text=Neon Rush")
	time.Sleep(1000 * time.Millisecond)

	// 3. Switch to Graph view
	log.Println("  🕸️ Graph view...")
	page.Click("text=Graph")
	time.Sleep(2500 * time.Millisecond)

	// Click on graph nodes
	page.Click("text=Chrome Echo")
	time.Sleep(1000 * time.Millisecond)

	// 4. Back to Library and toggle theme
	log.Println("  🌙 Theme toggle...")
	page.Click("text=Library")
	time.Sleep(1000 * time.Millisecond)
	page.Click("text=Dark")
	time.Sleep(1500 * time.Millisecond)

	// Toggle back to dark
	page.Click("text=Light")
	time.Sleep(1500 * time.Millisecond)

	// 5. Final view of Library
	page.Click("text=Berghain Sunrise")
	time.Sleep(1500 * time.Millisecond)

	// Close context to finalize video
	log.Println("  💾 Saving video...")
	videoPath, err := page.Video().Path()
	if err != nil {
		log.Printf("Warning: could not get video path: %v", err)
	}
	context.Close()

	// Also take static screenshots with a fresh context
	log.Println("📸 Taking static screenshots...")
	takeStaticScreenshots(browser)

	// Convert video to webp using ffmpeg + ImageMagick
	if videoPath != "" {
		log.Println("🎞️ Converting to animated WebP...")
		convertToWebp(videoPath, "docs/assets/screens/algiers-demo.webp")
	}

	log.Println("✅ Demo video and screenshots saved to docs/assets/screens/")
}

func takeStaticScreenshots(browser playwright.Browser) {
	page, err := browser.NewPage()
	if err != nil {
		log.Printf("Warning: could not create page for screenshots: %v", err)
		return
	}
	defer page.Close()

	// Set viewport size for high-res screenshots
	page.SetViewportSize(2560, 1440)

	if _, err = page.Goto("http://localhost:5173", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Printf("Warning: could not load page: %v", err)
		return
	}
	time.Sleep(2 * time.Second)

	// Hero shot
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("docs/assets/screens/algiers-hero.png"),
	})

	// Library view
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("docs/assets/screens/algiers-library-view.png"),
	})

	// Set Builder
	page.Click("text=Set Builder")
	time.Sleep(500 * time.Millisecond)
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("docs/assets/screens/algiers-set-builder.png"),
	})

	// Graph view
	page.Click("text=Graph")
	time.Sleep(800 * time.Millisecond)
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("docs/assets/screens/algiers-graph-view.png"),
	})

	// Light mode
	page.Click("text=Library")
	time.Sleep(300 * time.Millisecond)
	page.Click("text=Dark")
	time.Sleep(500 * time.Millisecond)
	page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("docs/assets/screens/algiers-light-mode.png"),
	})
}

func convertToWebp(inputPath, outputPath string) {
	tempGif := "docs/assets/screens/temp.gif"

	// Step 1: Convert video to optimized GIF using ffmpeg
	// - 10fps for smooth but small animation
	// - 960px width for reasonable file size
	// - 128 color palette for compression
	log.Println("  Creating optimized GIF...")
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", inputPath,
		"-vf", "fps=10,scale=960:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse",
		"-loop", "0",
		tempGif,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Warning: ffmpeg conversion failed: %v\nOutput: %s", err, string(output))
		return
	}

	// Step 2: Convert GIF to animated WebP using ImageMagick
	log.Println("  Converting to WebP...")
	cmd = exec.Command("magick",
		tempGif,
		"-quality", "70",
		"-define", "webp:lossless=false",
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Warning: ImageMagick conversion failed: %v\nOutput: %s", err, string(output))
		return
	}

	// Clean up temp files
	os.Remove(tempGif)
	os.Remove(inputPath)

	// Report file size
	if info, err := os.Stat(outputPath); err == nil {
		log.Printf("  Created %s (%.1f MB)", outputPath, float64(info.Size())/(1024*1024))
	}
}
