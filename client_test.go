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
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{
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
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{
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
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{
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
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{
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

	t.Run("invalid developer key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errorId":          1,
				"errorCode":        "ERROR_INVALID_DEVELOPER_KEY",
				"errorDescription": "Invalid developer key",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)
		client.SetDeveloperKey("DEV-invalid1234567")

		_, err := client.CreateTask(Task{"type": "ReCaptchaV2TaskProxyLess"})

		require.Error(t, err)
		var validationErr *ErrValidation
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("proxy connection failed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID:          1,
				ErrorCode:        "ERROR_PROXY_CONNECTION_FAILED",
				ErrorDescription: "Could not connect through proxy",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.CreateTask(Task{"type": "ReCaptchaV2Task"})

		require.Error(t, err)
		var validationErr *ErrValidation
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("proxy banned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{
				ErrorID:          1,
				ErrorCode:        "ERROR_PROXY_BANNED",
				ErrorDescription: "Proxy IP blocked by target",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.CreateTask(Task{"type": "ReCaptchaV2Task"})

		require.Error(t, err)
		var validationErr *ErrValidation
		assert.ErrorAs(t, err, &validationErr)
	})
}

func TestGetTaskResult(t *testing.T) {
	t.Run("processing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(TaskResult{
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
			_ = json.NewEncoder(w).Encode(TaskResult{
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
			_ = json.NewEncoder(w).Encode(TaskResult{
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
		_ = json.NewEncoder(w).Encode(GetBalanceResponse{
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
				_ = json.NewEncoder(w).Encode(CreateTaskResponse{
					ErrorID: 0,
					TaskID:  "test-task-id",
				})
			} else if r.URL.Path == "/getTaskResult" {
				callCount++
				if callCount == 1 {
					_ = json.NewEncoder(w).Encode(TaskResult{
						ErrorID: 0,
						Status:  "processing",
					})
				} else {
					_ = json.NewEncoder(w).Encode(TaskResult{
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
				_ = json.NewEncoder(w).Encode(CreateTaskResponse{
					ErrorID: 0,
					TaskID:  "test-task-id",
				})
			} else if r.URL.Path == "/getTaskResult" {
				_ = json.NewEncoder(w).Encode(TaskResult{
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
				_ = json.NewEncoder(w).Encode(CreateTaskResponse{
					ErrorID: 0,
					TaskID:  "test-task-id",
				})
			} else if r.URL.Path == "/getTaskResult" {
				_ = json.NewEncoder(w).Encode(TaskResult{
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
			_ = json.NewEncoder(w).Encode(GetBalanceResponse{
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

func TestGetPricing(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/pricing", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.Header.Get("User-Agent"), "capbypass-sdk-go")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(PricingResponse{
				Pricing: []PricingItem{
					{TaskType: TaskTypeReCaptchaV2ProxyLess, UserCost: 1.5, Status: "active"},
					{TaskType: TaskTypeHCaptcha, UserCost: 2.0, Status: "active"},
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		pricing, err := client.GetPricing()

		require.NoError(t, err)
		require.Len(t, pricing, 2)
		assert.Equal(t, TaskTypeReCaptchaV2ProxyLess, pricing[0].TaskType)
		assert.Equal(t, 1.5, pricing[0].UserCost)
		assert.Equal(t, "active", pricing[0].Status)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetPricing()

		require.Error(t, err)
		var parseErr *ErrParse
		assert.ErrorAs(t, err, &parseErr)
	})

	t.Run("server error 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetPricing()

		require.Error(t, err)
		var serverErr *ErrServer
		assert.ErrorAs(t, err, &serverErr)
	})

	t.Run("client error 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetPricing()

		require.Error(t, err)
		var netErr *ErrNetwork
		assert.ErrorAs(t, err, &netErr)
	})

	t.Run("connection failure", func(t *testing.T) {
		client, _ := NewClient("test-key")
		// Point at a closed port so the GET dial fails (covers makeGetRequest network path).
		client.SetBaseURL("http://127.0.0.1:1")

		_, err := client.GetPricing()

		require.Error(t, err)
		var netErr *ErrNetwork
		assert.ErrorAs(t, err, &netErr)
		assert.Contains(t, err.Error(), "Connection failed")
	})
}

func TestGetPricingGatewayRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PricingResponse{
			Pricing: []PricingItem{{TaskType: TaskTypeHCaptcha, UserCost: 2.0, Status: "active"}},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-key")
	client.SetBaseURL(server.URL)

	pricing, err := client.GetPricing()

	require.NoError(t, err)
	require.Len(t, pricing, 1)
	assert.Equal(t, 2, callCount) // initial 502 + 1 retry success
}

func TestGetBalanceErrors(t *testing.T) {
	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(GetBalanceResponse{
				ErrorID:          1,
				ErrorCode:        "ERROR_KEY_DOES_NOT_EXIST",
				ErrorDescription: "Account not found",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		balance, err := client.GetBalance()

		require.Error(t, err)
		assert.Equal(t, 0.0, balance)
		var authErr *ErrAuthentication
		assert.ErrorAs(t, err, &authErr)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{not-json"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetBalance()

		require.Error(t, err)
		var parseErr *ErrParse
		assert.ErrorAs(t, err, &parseErr)
	})
}

// TestParseErrorBranches covers the remaining parseError switch arms that
// createTask/getTaskResult tests don't already exercise.
func TestParseErrorBranches(t *testing.T) {
	cases := []struct {
		name   string
		code   string
		assert func(t *testing.T, err error)
	}{
		{
			name: "captcha unsolvable -> solver",
			code: "ERROR_CAPTCHA_UNSOLVABLE",
			assert: func(t *testing.T, err error) {
				var e *ErrSolver
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "timeout -> timeout",
			code: "ERROR_TIMEOUT",
			assert: func(t *testing.T, err error) {
				var e *ErrTimeout
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "proxy not defined -> validation",
			code: "ERROR_PROXY_NOT_DEFINED",
			assert: func(t *testing.T, err error) {
				var e *ErrValidation
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "wrong task type -> validation",
			code: "ERROR_WRONG_TASK_TYPE",
			assert: func(t *testing.T, err error) {
				var e *ErrValidation
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "task type coming soon -> validation",
			code: "ERROR_TASK_TYPE_COMING_SOON",
			assert: func(t *testing.T, err error) {
				var e *ErrValidation
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "task type inactive -> validation",
			code: "ERROR_TASK_TYPE_INACTIVE",
			assert: func(t *testing.T, err error) {
				var e *ErrValidation
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "queue full -> internal (default)",
			code: "ERROR_TASK_QUEUE_FULL",
			assert: func(t *testing.T, err error) {
				var e *ErrInternal
				assert.ErrorAs(t, err, &e)
			},
		},
		{
			name: "unknown code -> internal (default)",
			code: "ERROR_SOMETHING_NEW",
			assert: func(t *testing.T, err error) {
				var e *ErrInternal
				assert.ErrorAs(t, err, &e)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := parseError(tc.code, "desc for "+tc.code)
			require.Error(t, err)
			tc.assert(t, err)
		})
	}
}

// TestMakeRequestErrorBranches covers makeRequest non-retry HTTP error paths
// (429 rate limit, 500 server error, generic 4xx) and a non-JSON 200 body.
func TestMakeRequestErrorBranches(t *testing.T) {
	t.Run("rate limit 429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("slow down"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetBalance()

		require.Error(t, err)
		var rlErr *ErrRateLimit
		assert.ErrorAs(t, err, &rlErr)
	})

	t.Run("server 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("kaboom"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetBalance()

		require.Error(t, err)
		var serverErr *ErrServer
		assert.ErrorAs(t, err, &serverErr)
	})

	t.Run("generic 4xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("forbidden"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetBalance()

		require.Error(t, err)
		var netErr *ErrNetwork
		assert.ErrorAs(t, err, &netErr)
		assert.Contains(t, err.Error(), "HTTP 403")
	})

	t.Run("non-JSON 200 body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html>not json</html>"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key")
		client.SetBaseURL(server.URL)

		_, err := client.GetBalance()

		require.Error(t, err)
		var parseErr *ErrParse
		assert.ErrorAs(t, err, &parseErr)
	})

	t.Run("connection failure", func(t *testing.T) {
		client, _ := NewClient("test-key")
		client.SetBaseURL("http://127.0.0.1:1")

		_, err := client.GetBalance()

		require.Error(t, err)
		var netErr *ErrNetwork
		assert.ErrorAs(t, err, &netErr)
		assert.Contains(t, err.Error(), "Connection failed")
	})
}

// fakeClockClient returns a client whose clock only moves when the SDK sleeps.
// events records "sleep <d>" and "poll" in call order.
func fakeClockClient(t *testing.T, readyAfter int) (*Client, *[]string) {
	t.Helper()

	events := []string{}
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/createTask":
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{ErrorID: 0, TaskID: "test-task-id"})
		case "/getTaskResult":
			polls++
			events = append(events, "poll")
			if readyAfter > 0 && polls >= readyAfter {
				_ = json.NewEncoder(w).Encode(TaskResult{
					ErrorID:  0,
					Status:   "ready",
					Solution: map[string]interface{}{"token": "test"},
				})
			} else {
				_ = json.NewEncoder(w).Encode(TaskResult{ErrorID: 0, Status: "processing"})
			}
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(server.URL)

	clock := time.Unix(0, 0)
	client.now = func() time.Time { return clock }
	client.sleep = func(d time.Duration) {
		events = append(events, "sleep "+d.String())
		clock = clock.Add(d)
	}

	return client, &events
}

func TestPollDelay(t *testing.T) {
	want := []time.Duration{1, 1, 1, 2, 2, 3, 3, 3, 3}
	for i, w := range want {
		assert.Equal(t, w*time.Second, pollDelay(i+1), "wait before poll %d", i+1)
	}
}

func TestSolvePollSchedule(t *testing.T) {
	client, events := fakeClockClient(t, 8)

	solution, err := client.Solve(Task{"type": TaskTypeReCaptchaV2ProxyLess}, 120)
	require.NoError(t, err)
	assert.Equal(t, "test", solution["token"])

	// Every poll is preceded by its wait; the first wait comes before the first poll.
	assert.Equal(t, []string{
		"sleep 1s", "poll",
		"sleep 1s", "poll",
		"sleep 1s", "poll",
		"sleep 2s", "poll",
		"sleep 2s", "poll",
		"sleep 3s", "poll",
		"sleep 3s", "poll",
		"sleep 3s", "poll",
	}, *events)
}

func TestSolveTimeoutHonoured(t *testing.T) {
	// Waits 1,1,1,2,2 end at 1, 2, 3, 5 and 7 s. A 5 s timeout still allows
	// the poll at 5 s (elapsed is not above the timeout) and stops at the
	// wait that ends at 7 s: four polls.
	client, events := fakeClockClient(t, 0)

	_, err := client.Solve(Task{"type": TaskTypeReCaptchaV2ProxyLess}, 5)

	var timeoutErr *ErrTimeout
	require.ErrorAs(t, err, &timeoutErr)
	polls := 0
	for _, e := range *events {
		if e == "poll" {
			polls++
		}
	}
	assert.Equal(t, 4, polls)
}

func TestSolveDefaultTimeout(t *testing.T) {
	// timeout 0 means 120 s. After the 7 s ramp (1+1+1+2+2) the waits are 3 s:
	// the 37th ends at 118 s, the 38th at 121 s.
	client, events := fakeClockClient(t, 0)

	_, err := client.Solve(Task{"type": TaskTypeReCaptchaV2ProxyLess}, 0)

	var timeoutErr *ErrTimeout
	require.ErrorAs(t, err, &timeoutErr)
	polls := 0
	for _, e := range *events {
		if e == "poll" {
			polls++
		}
	}
	assert.Equal(t, 5+37, polls)
}
