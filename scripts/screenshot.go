package main

import (
	"log"
	"os"
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

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("new page: %v", err)
	}

	// Capture console messages
	page.On("console", func(msg playwright.ConsoleMessage) {
		log.Printf("[Browser Console] %s: %s", msg.Type(), msg.Text())
	})

	// High resolution for crisp screenshots
	page.SetViewportSize(2560, 1440)

	// Wait for app to fully load (dev server on 5173, preview on 4173)
	if _, err = page.Goto("http://localhost:5173", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("goto: %v", err)
	}

	// Wait for React to render with demo data
	time.Sleep(3 * time.Second)

	// Ensure docs/assets/screens exists
	os.MkdirAll("docs/assets/screens", 0755)

	// 1. Hero shot - Library view with populated data
	log.Println("📸 Capturing hero shot...")
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("docs/assets/screens/algiers-hero.png"),
		FullPage: playwright.Bool(false),
	}); err != nil {
		log.Fatalf("screenshot hero: %v", err)
	}

	// 2. Library view - same as hero but saved separately
	log.Println("📸 Capturing library view...")
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("docs/assets/screens/algiers-library-view.png"),
		FullPage: playwright.Bool(false),
	}); err != nil {
		log.Fatalf("screenshot library: %v", err)
	}

	// 3. Click on Set Builder tab
	log.Println("📸 Capturing Set Builder...")
	page.Click("text=Set Builder")
	time.Sleep(500 * time.Millisecond)
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("docs/assets/screens/algiers-set-builder.png"),
		FullPage: playwright.Bool(false),
	}); err != nil {
		log.Fatalf("screenshot set builder: %v", err)
	}

	// 4. Click on Graph tab
	log.Println("📸 Capturing Graph view...")
	page.Click("text=Graph")
	time.Sleep(800 * time.Millisecond)
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("docs/assets/screens/algiers-graph-view.png"),
		FullPage: playwright.Bool(false),
	}); err != nil {
		log.Fatalf("screenshot graph: %v", err)
	}

	// 5. Toggle to light mode
	log.Println("📸 Capturing light mode...")
	page.Click("text=Library")
	time.Sleep(300 * time.Millisecond)
	page.Click("text=Dark") // Click on the theme toggle
	time.Sleep(500 * time.Millisecond)
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String("docs/assets/screens/algiers-light-mode.png"),
		FullPage: playwright.Bool(false),
	}); err != nil {
		log.Fatalf("screenshot light: %v", err)
	}

	log.Println("✅ Screenshots saved to docs/assets/screens/")
}
