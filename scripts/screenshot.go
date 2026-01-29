package main

import (
    "context"
    "log"
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

    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()

    if _, err = page.Goto("http://localhost:4173", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
        log.Fatalf("goto: %v", err)
    }

    page.SetViewportSize(1440, 900)

    if _, err := page.Screenshot(playwright.PageScreenshotOptions{
        Path:     playwright.String("docs/assets/screens/algiers-hero.png"),
        FullPage: playwright.Bool(true),
    }); err != nil {
        log.Fatalf("screenshot hero: %v", err)
    }

    page.Evaluate(`window.scrollTo(0, document.body.scrollHeight * 0.6)`, nil)
    time.Sleep(500 * time.Millisecond)

    if _, err := page.Screenshot(playwright.PageScreenshotOptions{
        Path:     playwright.String("docs/assets/screens/algiers-set-builder.png"),
        FullPage: playwright.Bool(true),
    }); err != nil {
        log.Fatalf("screenshot set builder: %v", err)
    }

    <-ctx.Done()
}
