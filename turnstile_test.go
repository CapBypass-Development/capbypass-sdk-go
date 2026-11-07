package capbypass

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTurnstileTaskTypeConstants(t *testing.T) {
	assert.Equal(t, "TurnstileTask", TaskTypeTurnstile)
	assert.Equal(t, "TurnstileTaskProxyLess", TaskTypeTurnstileProxyLess)
}

// turnstileServer records the createTask body and answers a ready Turnstile solution.
func turnstileServer(t *testing.T, created *CreateTaskRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/createTask":
			require.NoError(t, json.NewDecoder(r.Body).Decode(created))
			_ = json.NewEncoder(w).Encode(CreateTaskResponse{TaskID: "ts-task"})
		case "/getTaskResult":
			_ = json.NewEncoder(w).Encode(TaskResult{
				Status: "ready",
				Solution: map[string]interface{}{
					"token":  "0.turnstile-token",
					"timing": map[string]interface{}{"solveMs": 6120},
				},
			})
		}
	}))
}

func TestSolveTurnstileProxyLess(t *testing.T) {
	var created CreateTaskRequest
	server := turnstileServer(t, &created)
	defer server.Close()

	client, _ := NewClient("test-key")
	client.SetBaseURL(server.URL)

	solution, err := client.Solve(Task{
		"type":       TaskTypeTurnstileProxyLess,
		"websiteURL": "https://example.com/login",
		"websiteKey": "0x4AAAAAAAxxxxxxxxxxxxxx",
	}, 120)

	require.NoError(t, err)
	assert.Equal(t, "0.turnstile-token", solution["token"])
	assert.Equal(t, float64(6120), solution["timing"].(map[string]interface{})["solveMs"])
	assert.Equal(t, "test-key", created.ClientKey)
	assert.Equal(t, Task{
		"type":       "TurnstileTaskProxyLess",
		"websiteURL": "https://example.com/login",
		"websiteKey": "0x4AAAAAAAxxxxxxxxxxxxxx",
	}, created.Task)
}

func TestSolveTurnstileWithProxyActionCData(t *testing.T) {
	var created CreateTaskRequest
	server := turnstileServer(t, &created)
	defer server.Close()

	client, _ := NewClient("test-key")
	client.SetBaseURL(server.URL)

	solution, err := client.Solve(Task{
		"type":       TaskTypeTurnstile,
		"websiteURL": "https://example.com/login",
		"websiteKey": "0x4AAAAAAAxxxxxxxxxxxxxx",
		"action":     "login",
		"cData":      "session-42",
		"proxy":      "http://user:pass@host:8080",
	}, 120)

	require.NoError(t, err)
	assert.Equal(t, "0.turnstile-token", solution["token"])
	assert.Equal(t, "TurnstileTask", created.Task["type"])
	assert.Equal(t, "login", created.Task["action"])
	assert.Equal(t, "session-42", created.Task["cData"])
	assert.Equal(t, "http://user:pass@host:8080", created.Task["proxy"])
}

func TestCreateTaskTurnstileGatewayErrors(t *testing.T) {
	cases := []struct {
		code string
		want error
	}{
		{"ERROR_INVALID_TASK_DATA", &ErrValidation{}},
		{"ERROR_TASK_TYPE_INACTIVE", &ErrValidation{}},
		{"ERROR_PROXY_FORMAT", &ErrValidation{}},
		{"ERROR_PROXY_BANNED", &ErrValidation{}},
		{"ERROR_BLOCKED_TARGET", &ErrValidation{}},
		{"ERROR_PROXY_CONNECTION_FAILED", &ErrValidation{}},
		{"ERROR_TASK_QUEUE_FULL", &ErrInternal{}},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(CreateTaskResponse{ErrorID: 1, ErrorCode: tc.code, ErrorDescription: "x"})
			}))
			defer server.Close()
			client, _ := NewClient("test-key")
			client.SetBaseURL(server.URL)

			_, err := client.CreateTask(Task{"type": TaskTypeTurnstile})
			assert.IsType(t, tc.want, err)
		})
	}
}
