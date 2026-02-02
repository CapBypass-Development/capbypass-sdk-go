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

	fmt.Println("Solving reCAPTCHA v3...")

	solution, err := client.Solve(capbypass.Task{
		"type":       capbypass.TaskTypeReCaptchaV3ProxyLess,
		"websiteURL": "https://example.com",
		"websiteKey": "6LcR_okUAAAAAPYrPe-HK_0RULO1aZM15ENyM-Mf",
		"pageAction": "submit",
		"minScore":   0.7,
	}, 120)

	if err != nil {
		log.Fatalf("✗ Error: %v", err)
	}

	fmt.Println("✓ CAPTCHA solved!")
	fmt.Printf("Token: %s...\n", solution["gRecaptchaResponse"].(string)[:80])
}
