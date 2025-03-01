package agent

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aoktiv/calculator_project/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) Post(url, contentType string, body string) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestFetchTask_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/task" {
			task := orchestrator.Task{ID: "1", Operation: "addition", Arg1: 2, Arg2: 3}
			response := struct {
				Task orchestrator.Task `json:"task"`
			}{Task: task}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer ts.Close()

	originalURL := "http://localhost:8080/internal/task"

	task, err := fetchTask()

	assert.NoError(t, err)
	assert.Equal(t, "1", task.ID)
	assert.Equal(t, "addition", task.Operation)
	assert.Equal(t, 2.0, task.Arg1)
	assert.Equal(t, 3.0, task.Arg2)
}

func TestCompute_Operation(t *testing.T) {
	tests := []struct {
		operation string
		arg1      float64
		arg2      float64
		expected  float64
	}{
		{"addition", 1, 1, 2},
		{"subtraction", 2, 1, 1},
		{"multiplication", 2, 3, 6},
		{"division", 6, 3, 2},
		{"unknown", 1, 1, math.NaN()},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			task := orchestrator.Task{
				Operation: tt.operation,
				Arg1:      tt.arg1,
				Arg2:      tt.arg2,
			}
			result := compute(task)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSendResult_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/internal/task/result", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var resultReq ResultRequest
		err := json.NewDecoder(r.Body).Decode(&resultReq)
		assert.NoError(t, err)

		assert.Equal(t, "1", resultReq.ID)
		assert.Equal(t, 5.0, resultReq.Result)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	originalPostURL := "http://localhost:8080/internal/task/result"

	err := sendResult("1", 5.0)

	assert.NoError(t, err)
}

func TestGetEnv(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "3")
	assert.Equal(t, 3, getEnv("COMPUTING_POWER", 2))

	// Тест без установленного значения
	os.Unsetenv("COMPUTING_POWER")
	assert.Equal(t, 2, getEnv("COMPUTING_POWER", 2))
}

func TestTaskWorker_FetchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Error fetching task", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &MockHTTPClient{}
	client.On("Get", "http://localhost:8080/internal/task").Return(nil, errors.New("Error fetching task"))

	_, err := fetchTask()
	assert.Error(t, err)

	client.AssertExpectations(t)
}

func TestTaskWorker_ComputeAndSendResult(t *testing.T) {
	taskServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := orchestrator.Task{ID: "1", Operation: "addition", Arg1: 2, Arg2: 3}
		response := struct {
			Task orchestrator.Task `json:"task"`
		}{Task: task}
		json.NewEncoder(w).Encode(response)
	}))
	defer taskServer.Close()

	resultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ResultRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "1", req.ID)
		assert.Equal(t, 5.0, req.Result)
		w.WriteHeader(http.StatusOK)
	}))
	defer resultServer.Close()
	
}
