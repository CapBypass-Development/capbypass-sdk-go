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

	fmt.Println("Solving AWS WAF challenge...")

	solution, err := client.Solve(capbypass.Task{
		"type":           capbypass.TaskTypeAntiAwsWafProxyLess,
		"websiteURL":     "https://login.tomorrowland.com",
		"awsChallengeJS": "https://b516434d791a.aa24f28d.eu-west-1.token.awswaf.com/b516434d791a/challenge.js",
	}, 120)

	if err != nil {
		log.Fatalf("✗ Error: %v", err)
	}

	fmt.Println("✓ AWS WAF challenge solved!")
	fmt.Printf("Cookie: %s\n", solution["cookie"])
}
