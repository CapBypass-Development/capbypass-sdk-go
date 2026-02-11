package capbypass

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	defaultBaseURL = "https://api.capbypass.pro"
	sdkVersion     = "1.0.0"
	userAgent      = "capbypass-sdk-go/" + sdkVersion
)

// Client is the CapBypass API client.
type Client struct {
	apiKey       string
	developerKey string
	baseURL      string
	httpClient   *http.Client
}

// NewClient creates a new CapBypass client.
// If apiKey is empty, it will be read from CAPBYPASS_API_KEY environment variable.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("CAPBYPASS_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required. Provide via constructor or CAPBYPASS_API_KEY env var")
	}

	return &Client{
		apiKey:       apiKey,
		developerKey: os.Getenv("CAPBYPASS_DEVELOPER_KEY"),
		baseURL:      defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SetBaseURL sets a custom base URL for the API.
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// SetDeveloperKey sets the developer affiliate key for commission attribution.
// Can also be set via CAPBYPASS_DEVELOPER_KEY environment variable.
func (c *Client) SetDeveloperKey(key string) {
	c.developerKey = key
}

// makeRequest makes an HTTP request with retry logic.
func (c *Client) makeRequest(endpoint string, payload interface{}, maxRetries int) ([]byte, error) {
	url := c.baseURL + endpoint
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, &ErrParse{Message: "Failed to marshal request", Err: err}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, &ErrNetwork{Message: "Failed to create request", Err: err}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = &ErrNetwork{Message: "Connection failed", Err: err}
			if attempt < maxRetries {
				backoff := time.Duration(math.Min(10, math.Pow(2, float64(attempt)))+rand.Float64()) * time.Second
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &ErrParse{Message: "Failed to read response", Err: err}
		}

		// Handle gateway errors with retry
		if resp.StatusCode == 502 || resp.StatusCode == 503 || resp.StatusCode == 504 {
			if attempt < maxRetries {
				backoff := time.Duration(math.Min(10, math.Pow(2, float64(attempt)))+rand.Float64()) * time.Second
				time.Sleep(backoff)
				continue
			}
			return nil, &ErrGateway{StatusCode: resp.StatusCode, Message: string(body)}
		}

		// Handle other HTTP errors
		if resp.StatusCode == 429 {
			return nil, &ErrRateLimit{Message: string(body)}
		}
		if resp.StatusCode == 500 {
			return nil, &ErrServer{StatusCode: resp.StatusCode, Message: string(body)}
		}
		if resp.StatusCode >= 400 {
			return nil, &ErrNetwork{Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
		}

		return body, nil
	}

	return nil, lastErr
}

// makeGetRequest makes an HTTP GET request with retry logic.
func (c *Client) makeGetRequest(endpoint string, maxRetries int) ([]byte, error) {
	url := c.baseURL + endpoint
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, &ErrNetwork{Message: "Failed to create request", Err: err}
		}

		req.Header.Set("User-Agent", userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = &ErrNetwork{Message: "Connection failed", Err: err}
			if attempt < maxRetries {
				backoff := time.Duration(math.Min(10, math.Pow(2, float64(attempt)))+rand.Float64()) * time.Second
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &ErrParse{Message: "Failed to read response", Err: err}
		}

		if resp.StatusCode == 502 || resp.StatusCode == 503 || resp.StatusCode == 504 {
			if attempt < maxRetries {
				backoff := time.Duration(math.Min(10, math.Pow(2, float64(attempt)))+rand.Float64()) * time.Second
				time.Sleep(backoff)
				continue
			}
			return nil, &ErrGateway{StatusCode: resp.StatusCode, Message: string(body)}
		}

		if resp.StatusCode == 500 {
			return nil, &ErrServer{StatusCode: resp.StatusCode, Message: string(body)}
		}

		if resp.StatusCode >= 400 {
			return nil, &ErrNetwork{Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
		}

		return body, nil
	}

	return nil, lastErr
}

// parseError parses API error response and returns appropriate error type.
func parseError(errorCode, errorDesc string) error {
	switch errorCode {
	case "ERROR_KEY_DOES_NOT_EXIST", "ERROR_KEY_DENIED_ACCESS":
		return newAuthenticationError(errorCode, errorDesc)
	case "ERROR_ZERO_BALANCE", "ERROR_NO_SLOT_AVAILABLE":
		return newInsufficientBalanceError(errorCode, errorDesc)
	case "ERROR_INVALID_TASK_DATA", "ERROR_TASK_ABSENT", "ERROR_TASK_NOT_SUPPORTED",
		"TASK_TYPE_COMING_SOON", "TASK_TYPE_INACTIVE":
		return newValidationError(errorCode, errorDesc)
	case "ERROR_TASK_NOT_FOUND":
		return newTaskNotFoundError(errorCode, errorDesc)
	default:
		return newInternalError(errorCode, errorDesc)
	}
}

// CreateTask creates a new CAPTCHA solving task.
func (c *Client) CreateTask(task Task) (string, error) {
	req := CreateTaskRequest{
		ClientKey:    c.apiKey,
		Task:         task,
		DeveloperKey: c.developerKey,
	}

	body, err := c.makeRequest("/createTask", req, 3)
	if err != nil {
		return "", err
	}

	var resp CreateTaskResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", &ErrParse{Message: "Failed to parse createTask response", Err: err}
	}

	if resp.ErrorID != 0 {
		return "", parseError(resp.ErrorCode, resp.ErrorDescription)
	}

	return resp.TaskID, nil
}

// GetTaskResult retrieves the result of a task.
func (c *Client) GetTaskResult(taskID string) (*TaskResult, error) {
	req := GetTaskResultRequest{
		ClientKey: c.apiKey,
		TaskID:    taskID,
	}

	body, err := c.makeRequest("/getTaskResult", req, 3)
	if err != nil {
		return nil, err
	}

	var resp TaskResult
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &ErrParse{Message: "Failed to parse getTaskResult response", Err: err}
	}

	if resp.ErrorID != 0 {
		return nil, parseError(resp.ErrorCode, resp.ErrorDescription)
	}

	return &resp, nil
}

// GetBalance retrieves the account balance.
func (c *Client) GetBalance() (float64, error) {
	req := GetBalanceRequest{
		ClientKey: c.apiKey,
	}

	body, err := c.makeRequest("/getBalance", req, 3)
	if err != nil {
		return 0, err
	}

	var resp GetBalanceResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, &ErrParse{Message: "Failed to parse getBalance response", Err: err}
	}

	if resp.ErrorID != 0 {
		return 0, parseError(resp.ErrorCode, resp.ErrorDescription)
	}

	return resp.Balance, nil
}

// GetPricing retrieves pricing for all task types.
// This is a public endpoint and does not require authentication.
func (c *Client) GetPricing() ([]PricingItem, error) {
	body, err := c.makeGetRequest("/pricing", 3)
	if err != nil {
		return nil, err
	}

	var resp PricingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &ErrParse{Message: "Failed to parse pricing response", Err: err}
	}

	return resp.Pricing, nil
}

// Solve creates a task and polls until it's solved or times out.
func (c *Client) Solve(task Task, timeout int) (map[string]interface{}, error) {
	if timeout == 0 {
		timeout = 120
	}

	taskID, err := c.CreateTask(task)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	attempt := 0

	for {
		elapsed := time.Since(startTime).Seconds()
		if elapsed > float64(timeout) {
			return nil, newTimeoutError()
		}

		result, err := c.GetTaskResult(taskID)
		if err != nil {
			return nil, err
		}

		if result.Status == "ready" {
			return result.Solution, nil
		}

		if result.Status == "failed" {
			errDesc := result.ErrorDescription
			if errDesc == "" {
				errDesc = "Task failed"
			}
			return nil, newSolverError(errDesc)
		}

		// Adaptive polling: min(5, ceil(attempt / 2))
		attempt++
		pollInterval := int(math.Min(5, math.Ceil(float64(attempt)/2)))
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}
