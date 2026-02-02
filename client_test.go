package capbypass

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("with API key", func(t *testing.T) {
		client, err := NewClient("test-key")
		require.NoError(t, err)
		assert.Equal(t, "test-key", client.apiKey)
		assert.Equal(t, defaultBaseURL, client.baseURL)
	})

	t.Run("from environment variable", func(t *testing.T) {
		os.Setenv("CAPBYPASS_API_KEY", "env-key")
		defer os.Unsetenv("CAPBYPASS_API_KEY")

		client, err := NewClient("")
		require.NoError(t, err)
		assert.Equal(t, "env-key", client.apiKey)
	})

	t.Run("parameter takes priority over env", func(t *testing.T) {
		os.Setenv("CAPBYPASS_API_KEY", "env-key")
		defer os.Unsetenv("CAPBYPASS_API_KEY")

		client, err := NewClient("param-key")
		require.NoError(t, err)
		assert.Equal(t, "param-key", client.apiKey)
	})

	t.Run("no API key", func(t *testing.T) {
		client, err := NewClient("")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "API key is required")
	})
}

func TestCreateTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/createTask", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("User-Agent"), "capbypass-sdk-go")

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID: 0,
				TaskID:  "test-task-id-123",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		taskID, err := client.CreateTask(Task{
			"type":       TaskTypeReCaptchaV2ProxyLess,
			"websiteURL": "https://example.com",
			"websiteKey": "test-key",
		})

		require.NoError(t, err)
		assert.Equal(t, "test-task-id-123", taskID)
	})

	t.Run("invalid API key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID:          1,
				ErrorCode:        "ERROR_KEY_DOES_NOT_EXIST",
				ErrorDescription: "Account not found",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		taskID, err := client.CreateTask(Task{"type": TaskTypeReCaptchaV2ProxyLess})

		require.Error(t, err)
		assert.Empty(t, taskID)
		var authErr *ErrAuthentication
		assert.ErrorAs(t, err, &authErr)
		assert.Contains(t, err.Error(), "Account not found")
	})

	t.Run("zero balance", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID:          1,
				ErrorCode:        "ERROR_ZERO_BALANCE",
				ErrorDescription: "Insufficient balance",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.CreateTask(Task{"type": TaskTypeReCaptchaV2ProxyLess})

		require.Error(t, err)
		var balanceErr *ErrInsufficientBalance
		assert.ErrorAs(t, err, &balanceErr)
	})

	t.Run("invalid task data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID:          1,
				ErrorCode:        "ERROR_INVALID_TASK_DATA",
				ErrorDescription: "Invalid task type",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.CreateTask(Task{"type": "InvalidTaskType"})

		require.Error(t, err)
		var validationErr *ErrValidation
		assert.ErrorAs(t, err, &validationErr)
	})
}

func TestGetTaskResult(t *testing.T) {
	t.Run("processing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TaskResult{
				ErrorID: 0,
				Status:  "processing",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		result, err := client.GetTaskResult("test-task-id")

		require.NoError(t, err)
		assert.Equal(t, "processing", result.Status)
	})

	t.Run("ready", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TaskResult{
				ErrorID: 0,
				Status:  "ready",
				Solution: map[string]interface{}{
					"gRecaptchaResponse": "test-token",
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		result, err := client.GetTaskResult("test-task-id")

		require.NoError(t, err)
		assert.Equal(t, "ready", result.Status)
		assert.Equal(t, "test-token", result.Solution["gRecaptchaResponse"])
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TaskResult{
				ErrorID:          16,
				ErrorCode:        "ERROR_TASK_NOT_FOUND",
				ErrorDescription: "Task not found",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetTaskResult("invalid-task-id")

		require.Error(t, err)
		var notFoundErr *ErrTaskNotFound
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestGetBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/getBalance", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GetBalanceResponse{
			ErrorID: 0,
			Balance: 42.5,
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-key")
	client.SetBaseURL(server.URL)

	balance, err := client.GetBalance()

	require.NoError(t, err)
	assert.Equal(t, 42.5, balance)
}

func TestSolve(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			if r.URL.Path == "/createTask" {
				json.NewEncoder(w).Encode(CreateTaskResponse{
					ErrorID: 0,
					TaskID:  "test-task-id",
				})
			} else if r.URL.Path == "/getTaskResult" {
				callCount++
				if callCount == 1 {
					json.NewEncoder(w).Encode(TaskResult{
						ErrorID: 0,
						Status:  "processing",
					})
				} else {
					json.NewEncoder(w).Encode(TaskResult{
						ErrorID: 0,
						Status:  "ready",
						Solution: map[string]interface{}{
							"gRecaptchaResponse": "solved-token",
						},
					})
				}
			}
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		solution, err := client.Solve(Task{
			"type":       TaskTypeReCaptchaV2ProxyLess,
			"websiteURL": "https://example.com",
			"websiteKey": "test-key",
		}, 120)

		require.NoError(t, err)
		assert.Equal(t, "solved-token", solution["gRecaptchaResponse"])
	})

	t.Run("failed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			if r.URL.Path == "/createTask" {
				json.NewEncoder(w).Encode(CreateTaskResponse{
					ErrorID: 0,
					TaskID:  "test-task-id",
				})
			} else if r.URL.Path == "/getTaskResult" {
				json.NewEncoder(w).Encode(TaskResult{
					ErrorID:          0,
					Status:           "failed",
					ErrorDescription: "CAPTCHA unsolvable",
				})
			}
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.Solve(Task{"type": TaskTypeReCaptchaV2ProxyLess}, 120)

		require.Error(t, err)
		var solverErr *ErrSolver
		assert.ErrorAs(t, err, &solverErr)
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			if r.URL.Path == "/createTask" {
				json.NewEncoder(w).Encode(CreateTaskResponse{
					ErrorID: 0,
					TaskID:  "test-task-id",
				})
			} else if r.URL.Path == "/getTaskResult" {
				json.NewEncoder(w).Encode(TaskResult{
					ErrorID: 0,
					Status:  "processing",
				})
			}
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.Solve(Task{"type": TaskTypeReCaptchaV2ProxyLess}, 3)

		require.Error(t, err)
		var timeoutErr *ErrTimeout
		assert.ErrorAs(t, err, &timeoutErr)
	})
}

func TestRetryLogic(t *testing.T) {
	t.Run("gateway error retry", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GetBalanceResponse{
				ErrorID: 0,
				Balance: 10.0,
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		balance, err := client.GetBalance()

		require.NoError(t, err)
		assert.Equal(t, 10.0, balance)
		assert.Equal(t, 2, callCount)
	})

	t.Run("gateway error max retries", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetBalance()

		require.Error(t, err)
		var gatewayErr *ErrGateway
		assert.ErrorAs(t, err, &gatewayErr)
		assert.Equal(t, 4, callCount) // Initial + 3 retries
	})
}

func TestAdaptivePolling(t *testing.T) {
	callTimes := []time.Time{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/createTask" {
			json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID: 0,
				TaskID:  "test-task-id",
			})
		} else if r.URL.Path == "/getTaskResult" {
			callTimes = append(callTimes, time.Now())
			if len(callTimes) >= 7 {
				json.NewEncoder(w).Encode(TaskResult{
					ErrorID: 0,
					Status:  "ready",
					Solution: map[string]interface{}{
						"token": "test",
					},
				})
			} else {
				json.NewEncoder(w).Encode(TaskResult{
					ErrorID: 0,
					Status:  "processing",
				})
			}
		}
	}))
	defer server.Close()

	client, _ := NewClient("test-key")
	client.SetBaseURL(server.URL)

	_, err := client.Solve(Task{"type": TaskTypeReCaptchaV2ProxyLess}, 120)
	require.NoError(t, err)

	// Verify polling intervals: 1s, 1s, 2s, 2s, 3s, 3s
	for i := 1; i < len(callTimes); i++ {
		interval := callTimes[i].Sub(callTimes[i-1]).Seconds()
		var expectedInterval float64
		if i <= 2 {
			expectedInterval = 1.0
		} else if i <= 4 {
			expectedInterval = 2.0
		} else {
			expectedInterval = 3.0
		}
		assert.InDelta(t, expectedInterval, interval, 0.5, "Polling interval %d", i)
	}
}
