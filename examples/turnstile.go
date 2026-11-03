//go:build ignore

package main

import (
	"fmt"
	"log"

	"github.com/CapBypass-Development/capbypass-sdk-go"
)

func main() {
	client, err := capbypass.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Solving Cloudflare Turnstile...")

	solution, err := client.Solve(capbypass.Task{
		"type":       capbypass.TaskTypeTurnstileProxyLess,
		"websiteURL": "https://example.com/login",
		"websiteKey": "0x4AAAAAAAxxxxxxxxxxxxxx",
		"action":     "login",      // optional
		"cData":      "session-42", // optional
	}, 120)

	if err != nil {
		log.Fatalf("✗ Error: %v", err)
	}

	fmt.Println("✓ CAPTCHA solved!")
	fmt.Printf("Token: %s...\n", solution["token"].(string)[:50])
}
