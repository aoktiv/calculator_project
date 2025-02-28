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
	taskQueue      = make(chan Task, 100)
	taskMutex      sync.Mutex
	computeTimeout = 500 * time.Millisecond
)

type Expression struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Result     float64 `json:"result"`
	Expression string  `json:"expression"`
}

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

type ResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

func main() {
	computeTimeout = time.Duration(getEnv("TIME_ADDITION_MS", 500))

	// Initialize routes
	http.HandleFunc("/api/v1/calculate", calculateExpression)
	http.HandleFunc("/api/v1/expressions", getExpressions)
	http.HandleFunc("/api/v1/expressions/", getExpressionByID)
	http.HandleFunc("/internal/task", getTask)
	http.HandleFunc("/internal/task/result", acceptResult)
	go taskWorker()

	log.Println("Orchestrator started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func calculateExpression(w http.ResponseWriter, r *http.Request) {
	var expr struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil || expr.Expression == "" {
		http.Error(w, "Invalid data", http.StatusUnprocessableEntity)
		return
	}

	id := strconv.Itoa(len(expressions) + 1)

	expressions[id] = Expression{
		ID:         id,
		Status:     "pending",
		Expression: expr.Expression,
	}

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
	for {
		task := Task{
			ID:        "task1",
			Arg1:      5,
			Arg2:      10,
			Operation: "addition",
		}

		taskQueue <- task
		time.Sleep(1 * time.Second)
	}
}

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
