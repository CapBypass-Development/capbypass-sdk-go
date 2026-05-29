//go:build ignore

// reCAPTCHA v3 Detection Example (Go)
//
// Demonstrates how to programmatically detect whether a site uses
// reCAPTCHA v3 Standard or Enterprise, and automatically select
// the correct task type.
//
// This file carries the `ignore` build tag so it is excluded from
// `go build ./...` (it pulls in chromedp, which is not a dependency of
// the SDK itself). Run it directly once you have the extra dep:
//
//	go get github.com/chromedp/chromedp
//	go run examples/recaptcha_v3_detection.go

package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/CapBypass-Development/capbypass-sdk-go"
	"github.com/chromedp/chromedp"
)

// detectRecaptchaType detects whether a site uses reCAPTCHA v3 Standard or Enterprise
func detectRecaptchaType(targetURL string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var detectedType string

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`
            (() => {
                if (window.grecaptcha?.enterprise) return 'enterprise';
                if (window.grecaptcha) return 'standard';
                return 'unknown';
            })()
        `, &detectedType),
	)

	if err != nil {
		return "", fmt.Errorf("detection failed: %w", err)
	}

	if detectedType == "unknown" {
		return "", fmt.Errorf("could not detect reCAPTCHA type (neither standard nor enterprise found)")
	}

	fmt.Printf("✓ Detected: reCAPTCHA v3 %s\n", detectedType)
	return detectedType, nil
}

// solveWithAutoDetection detects the reCAPTCHA type and solves with the correct task type
func solveWithAutoDetection(client *capbypass.Client, targetURL, siteKey, action string) (capbypass.TaskResult, error) {
	fmt.Printf("\nDetecting reCAPTCHA type for: %s\n", targetURL)

	detectedType, err := detectRecaptchaType(targetURL)
	if err != nil {
		return capbypass.TaskResult{}, err
	}

	var taskType string
	if detectedType == "enterprise" {
		taskType = capbypass.TaskTypeReCaptchaV3EnterpriseProxyLess
	} else {
		taskType = capbypass.TaskTypeReCaptchaV3ProxyLess
	}

	fmt.Printf("Using task type: %s\n\n", taskType)

	solution, err := client.Solve(capbypass.Task{
		"type":       taskType,
		"websiteURL": targetURL,
		"websiteKey": siteKey,
		"pageAction": action,
	}, 120)

	return solution, err
}

// RecaptchaTypeCache caches detection results to avoid repeated browser launches
type RecaptchaTypeCache struct {
	cache map[string]string
}

// NewRecaptchaTypeCache creates a new cache instance
func NewRecaptchaTypeCache() *RecaptchaTypeCache {
	return &RecaptchaTypeCache{
		cache: make(map[string]string),
	}
}

// GetType gets the reCAPTCHA type for a URL, using cache if available
func (c *RecaptchaTypeCache) GetType(targetURL string) (string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	domain := parsedURL.Hostname()

	if cachedType, exists := c.cache[domain]; exists {
		fmt.Printf("Cache hit for %s → %s\n", domain, cachedType)
		return cachedType, nil
	}

	fmt.Printf("Cache miss for %s, detecting...\n", domain)
	detectedType, err := detectRecaptchaType(targetURL)
	if err != nil {
		return "", err
	}

	c.cache[domain] = detectedType
	fmt.Printf("Cached %s → %s\n", domain, detectedType)

	return detectedType, nil
}

// Clear clears the cache for a specific domain or the entire cache
func (c *RecaptchaTypeCache) Clear(domain string) {
	if domain == "" {
		c.cache = make(map[string]string)
	} else {
		delete(c.cache, domain)
	}
}

// ── Usage Examples ───────────────────────────────────────────────────────

func example1BasicDetection(client *capbypass.Client) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("Example 1: Basic Detection")
	fmt.Println("═══════════════════════════════════════════════════════\n")

	solution, err := solveWithAutoDetection(
		client,
		"https://example.com",
		"6Lc...",
		"submit",
	)

	if err != nil {
		fmt.Printf("✗ Failed: %v\n\n", err)
		return
	}

	token := solution["gRecaptchaResponse"].(string)
	fmt.Printf("✓ Token generated: %s...\n\n", token[:50])
}

func example2WithCaching(client *capbypass.Client) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("Example 2: Detection with Caching")
	fmt.Println("═══════════════════════════════════════════════════════\n")

	cache := NewRecaptchaTypeCache()

	// First solve - detects and caches
	type1, err := cache.GetType("https://example.com")
	if err != nil {
		fmt.Printf("✗ Failed: %v\n\n", err)
		return
	}

	taskType1 := capbypass.TaskTypeReCaptchaV3ProxyLess
	if type1 == "enterprise" {
		taskType1 = capbypass.TaskTypeReCaptchaV3EnterpriseProxyLess
	}

	solution1, err := client.Solve(capbypass.Task{
		"type":       taskType1,
		"websiteURL": "https://example.com",
		"websiteKey": "6Lc...",
		"pageAction": "submit",
	}, 120)

	if err != nil {
		fmt.Printf("✗ First solve failed: %v\n\n", err)
		return
	}

	token1 := solution1["gRecaptchaResponse"].(string)
	fmt.Printf("✓ First solve: %s...\n\n", token1[:50])

	// Second solve - uses cache (faster!)
	type2, err := cache.GetType("https://example.com")
	if err != nil {
		fmt.Printf("✗ Failed: %v\n\n", err)
		return
	}

	taskType2 := capbypass.TaskTypeReCaptchaV3ProxyLess
	if type2 == "enterprise" {
		taskType2 = capbypass.TaskTypeReCaptchaV3EnterpriseProxyLess
	}

	solution2, err := client.Solve(capbypass.Task{
		"type":       taskType2,
		"websiteURL": "https://example.com",
		"websiteKey": "6Lc...",
		"pageAction": "checkout",
	}, 120)

	if err != nil {
		fmt.Printf("✗ Second solve failed: %v\n\n", err)
		return
	}

	token2 := solution2["gRecaptchaResponse"].(string)
	fmt.Printf("✓ Second solve (cached): %s...\n\n", token2[:50])
}

func example3MultiSite(client *capbypass.Client) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("Example 3: Multiple Sites")
	fmt.Println("═══════════════════════════════════════════════════════\n")

	sites := []struct {
		url     string
		siteKey string
		action  string
	}{
		{"https://site1.com", "6Lc...", "login"},
		{"https://site2.com", "6Ld...", "submit"},
		{"https://site3.com", "6Le...", "checkout"},
	}

	for _, site := range sites {
		fmt.Printf("\nProcessing: %s\n", site.url)

		solution, err := solveWithAutoDetection(client, site.url, site.siteKey, site.action)
		if err != nil {
			fmt.Printf("  ✗ Failed: %v\n", err)
			continue
		}

		token := solution["gRecaptchaResponse"].(string)
		fmt.Printf("  ✓ Success: %s...\n", token[:50])
	}

	fmt.Println()
}

func main() {
	apiKey := os.Getenv("CAPBYPASS_API_KEY")
	if apiKey == "" {
		log.Fatal("ERROR: CAPBYPASS_API_KEY environment variable not set")
	}

	client, err := capbypass.NewClient(apiKey)
	if err != nil {
		log.Fatalf("failed to init client: %v", err)
	}

	example1BasicDetection(client)
	example2WithCaching(client)
	example3MultiSite(client)
}
