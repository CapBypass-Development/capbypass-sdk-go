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

	fmt.Println("Solving reCAPTCHA v2...")

	solution, err := client.Solve(capbypass.Task{
		"type":       capbypass.TaskTypeReCaptchaV2ProxyLess,
		"websiteURL": "https://www.google.com/recaptcha/api2/demo",
		"websiteKey": "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
	}, 120)

	if err != nil {
		log.Fatalf("✗ Error: %v", err)
	}

	fmt.Println("✓ CAPTCHA solved!")
	fmt.Printf("Token: %s...\n", solution["gRecaptchaResponse"].(string)[:80])
}
