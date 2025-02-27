package orchestrator

import (
	"encoding/json"
	_ "fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	expressions    = make(map[string]Expression)
	taskQueue      = make(chan Task, 100) // Queue to store tasks
	taskMutex      sync.Mutex
	computeTimeout = 500 * time.Millisecond // Default timeout
)

// Expression represents the arithmetic expression
type Expression struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Result     float64 `json:"result"`
	Expression string  `json:"expression"`
}

// Task represents the task for the agent to process
type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

// ResultRequest represents the result from the agent
type ResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

func main() {
	// Get compute time from environment variables
	computeTimeout = time.Duration(getEnv("TIME_ADDITION_MS", 500))

	// Initialize routes
	http.HandleFunc("/api/v1/calculate", calculateExpression)
	http.HandleFunc("/api/v1/expressions", getExpressions)
	http.HandleFunc("/api/v1/expressions/", getExpressionByID)
	http.HandleFunc("/internal/task", getTask)
	http.HandleFunc("/internal/task/result", acceptResult)

	// Start a background worker to handle tasks
	go taskWorker()

	// Start server
	log.Println("Orchestrator started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func calculateExpression(w http.ResponseWriter, r *http.Request) {
	var expr struct {
		Expression string `json:"expression"`
	}
	// Parse the request body
	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil || expr.Expression == "" {
		http.Error(w, "Invalid data", http.StatusUnprocessableEntity)
		return
	}

	// Generate unique ID for the expression
	id := strconv.Itoa(len(expressions) + 1)

	// Add expression to the map
	expressions[id] = Expression{
		ID:         id,
		Status:     "pending",
		Expression: expr.Expression,
	}

	// Respond with the ID of the expression
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func getExpressions(w http.ResponseWriter, r *http.Request) {
	var exprList []Expression
	for _, expr := range expressions {
		exprList = append(exprList, expr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": exprList})
}

func getExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/expressions/"):]

	expr, found := expressions[id]
	if !found {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]Expression{"expression": expr})
}

func getTask(w http.ResponseWriter, r *http.Request) {
	// Wait until there's a task available in the task queue
	taskMutex.Lock()
	task := <-taskQueue
	taskMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]Task{"task": task})
}

func acceptResult(w http.ResponseWriter, r *http.Request) {
	var result ResultRequest
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil || result.ID == "" {
		http.Error(w, "Invalid data", http.StatusUnprocessableEntity)
		return
	}

	// Update the expression status with the result
	expr, found := expressions[result.ID]
	if !found {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	expr.Status = "completed"
	expr.Result = result.Result
	expressions[result.ID] = expr

	w.WriteHeader(http.StatusOK)
}

func taskWorker() {
	// Periodically check for expressions to compute
	for {
		// Retrieve task from the task queue (simplified)
		task := Task{
			ID:        "task1",
			Arg1:      5,
			Arg2:      10,
			Operation: "addition",
		}

		taskQueue <- task
		time.Sleep(1 * time.Second) // Sleep before getting new task
	}
}

// Helper function to get environment variable with default
func getEnv(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return v
}
