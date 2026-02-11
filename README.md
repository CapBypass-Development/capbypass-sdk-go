# CapBypass Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/CapBypass-Development/capbypass-sdk-go.svg)](https://pkg.go.dev/github.com/CapBypass-Development/capbypass-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/CapBypass-Development/capbypass-sdk-go)](https://goreportcard.com/report/github.com/CapBypass-Development/capbypass-sdk-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go SDK for the CapBypass CAPTCHA solving service. Supports reCAPTCHA v2, reCAPTCHA v3, and AWS WAF challenges.

## Features

- ✅ **Simple API**: One-line `Solve()` method or advanced `CreateTask()`/`GetTaskResult()` control
- 🔄 **Automatic Polling**: Built-in adaptive polling with exponential backoff
- 🛡️ **Robust Error Handling**: Typed errors for all API and network failures
- 🔁 **Smart Retry Logic**: Automatic retry on network/gateway errors
- 🎯 **Type-Safe**: Full Go type safety with structs and constants

## Installation

```bash
go get github.com/CapBypass-Development/capbypass-sdk-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/CapBypass-Development/capbypass-sdk-go"
)

func main() {
    client, err := capbypass.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    solution, err := client.Solve(capbypass.Task{
        "type":       capbypass.TaskTypeReCaptchaV2ProxyLess,
        "websiteURL": "https://www.google.com/recaptcha/api2/demo",
        "websiteKey": "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
    }, 120)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Token:", solution["gRecaptchaResponse"])
}
```

## API Reference

### Client Creation

```go
// With API key parameter
client, err := capbypass.NewClient("your-api-key")

// From CAPBYPASS_API_KEY environment variable
client, err := capbypass.NewClient("")
```

### Simple API (Recommended)

**Solve** - One-step CAPTCHA solving:

```go
solution, err := client.Solve(task, timeout)
```

- `task`: Task configuration (see Task Types below)
- `timeout`: Maximum wait time in seconds (default: 120)
- Returns: Solution map or error

### Advanced API

For full control over task lifecycle:

```go
// Create task
taskID, err := client.CreateTask(task)

// Poll for result
result, err := client.GetTaskResult(taskID)

// Check balance
balance, err := client.GetBalance()
```

## Task Types

### reCAPTCHA v2

```go
client.Solve(capbypass.Task{
    "type":       capbypass.TaskTypeReCaptchaV2ProxyLess,
    "websiteURL": "https://example.com",
    "websiteKey": "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
}, 120)
```

**Invisible reCAPTCHA v2:**

```go
client.Solve(capbypass.Task{
    "type":        capbypass.TaskTypeReCaptchaV2ProxyLess,
    "websiteURL":  "https://example.com",
    "websiteKey":  "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
    "isInvisible": true,
}, 120)
```

### reCAPTCHA v3

```go
client.Solve(capbypass.Task{
    "type":       capbypass.TaskTypeReCaptchaV3ProxyLess,
    "websiteURL": "https://example.com",
    "websiteKey": "6LcR_okUAAAAAPYrPe-HK_0RULO1aZM15ENyM-Mf",
    "pageAction": "submit",
}, 120)
```

### AWS WAF Challenge

```go
client.Solve(capbypass.Task{
    "type":           capbypass.TaskTypeAntiAwsWafProxyLess,
    "websiteURL":     "https://example.com",
    "awsChallengeJS": "https://[...].awswaf.com/[...]/challenge.js",
}, 120)
```

### With Proxy

All task types support proxy configuration:

```go
client.Solve(capbypass.Task{
    "type":          capbypass.TaskTypeReCaptchaV2Task,
    "websiteURL":    "https://example.com",
    "websiteKey":    "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
    "proxyType":     "http",
    "proxyAddress":  "proxy.example.com",
    "proxyPort":     8080,
    "proxyLogin":    "username",
    "proxyPassword": "password",
}, 120)
```

## Error Handling

The SDK uses typed errors for precise error handling:

```go
solution, err := client.Solve(task, 120)
if err != nil {
    switch e := err.(type) {
    case *capbypass.ErrAuthentication:
        // Invalid API key
    case *capbypass.ErrInsufficientBalance:
        // No balance
    case *capbypass.ErrValidation:
        // Invalid task parameters
    case *capbypass.ErrTimeout:
        // Task took too long
    case *capbypass.ErrSolver:
        // CAPTCHA could not be solved
    case *capbypass.ErrNetwork:
        // Network/connection error
    case *capbypass.ErrGateway:
        // Gateway error (502/503/504)
    default:
        // Other error
    }
}
```

## Task Type Constants

```go
// AWS WAF
TaskTypeAntiAwsWaf          = "AntiAwsWafTask"
TaskTypeAntiAwsWafProxyLess = "AntiAwsWafTaskProxyLess"

// reCAPTCHA v2
TaskTypeReCaptchaV2          = "ReCaptchaV2Task"
TaskTypeReCaptchaV2ProxyLess = "ReCaptchaV2TaskProxyLess"

// reCAPTCHA v3
TaskTypeReCaptchaV3          = "ReCaptchaV3Task"
TaskTypeReCaptchaV3ProxyLess = "ReCaptchaV3TaskProxyLess"

// reCAPTCHA v3 Enterprise
TaskTypeReCaptchaV3Enterprise          = "ReCaptchaV3EnterpriseTask"
TaskTypeReCaptchaV3EnterpriseProxyLess = "ReCaptchaV3EnterpriseTaskProxyLess"
```

## Documentation

### 📚 Core Documentation
- [Quick Start Guide](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/quickstart/go.md)
- [Complete API Reference](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/api-reference/go-sdk.md)
- [Go Package Documentation](https://pkg.go.dev/github.com/CapBypass-Development/capbypass-sdk-go)

### 🔧 Advanced Guides
- [Proxy Configuration](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/guides/proxy-configuration.md) — HTTP, HTTPS, SOCKS5 proxy support with rotation strategies
- [Error Handling](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/guides/error-handling.md) — Retry strategies, circuit breakers, production alerting
- [Performance Optimization](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/guides/performance-optimization.md) — Concurrent solving, connection pooling, token caching
- [Production Deployment](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/guides/production-deployment.md) — Kubernetes, AWS Lambda, monitoring, security

### 🔄 Migration
- [Migrating from Capsolver](https://github.com/CapBypass-Development/capbypass-sdks/blob/main/docs/migration/from-capsolver.md) — 100% API compatible, drop-in replacement

## Examples

### Basic Examples
See the [examples](examples/) directory for complete runnable examples:
- [recaptcha_v2.go](examples/recaptcha_v2.go) - reCAPTCHA v2 solving
- [recaptcha_v3.go](examples/recaptcha_v3.go) - reCAPTCHA v3 solving
- [aws_waf.go](examples/aws_waf.go) - AWS WAF challenge solving

### Advanced Examples
Full integration examples in the [documentation](https://github.com/CapBypass-Development/capbypass-sdks/tree/main/docs/examples):
- E-commerce checkout automation
- Social media automation
- Web scraping with CAPTCHA handling
- Microservice integration patterns

## Testing

```bash
# Run unit tests
go test -v

# Run with coverage
go test -v -cover

# Run integration tests (requires API key)
export CAPBYPASS_API_KEY=your-api-key
go test -v -tags=integration
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- [Documentation](https://capbypass.dev/docs/sdks/go)
- [API Reference](https://pkg.go.dev/github.com/CapBypass-Development/capbypass-sdk-go)
- [GitHub Repository](https://github.com/CapBypass-Development/capbypass-sdk-go)
- [Bug Reports](https://github.com/CapBypass-Development/capbypass-sdk-go/issues)
