package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aoktiv/calculator_project/orchestrator"
)

var computingPower = 2 // Number of worker goroutines for agent

// ResultRequest represents the result from the agent
type ResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

func main() {
	// Get computing power from environment variable
	computingPower = getEnv("COMPUTING_POWER", 2)

	// Start worker goroutines
	for i := 0; i < computingPower; i++ {
		go taskWorker()
	}

	// Block forever
	select {}
}

func taskWorker() {
	for {
		// Fetch task from orchestrator
		task, err := fetchTask()
		if err != nil {
			log.Println("Error fetching task:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Compute the result
		result := compute(task)

		// Send result back to orchestrator
		err = sendResult(task.ID, result)
		if err != nil {
			log.Println("Error sending result:", err)
		}

		time.Sleep(1 * time.Second) // Simulate delay
	}
}

func fetchTask() (Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Task{}, fmt.Errorf("failed to get task, status: %d", resp.StatusCode)
	}

	var result struct {
		Task Task `json:"task"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return Task{}, err
	}

	return result.Task, nil
}

func compute(task Task) float64 {
	var result float64
	switch task.Operation {
	case "addition":
		result = task.Arg1 + task.Arg2
	case "subtraction":
		result = task.Arg1 - task.Arg2
	case "multiplication":
		result = task.Arg1 * task.Arg2
	case "division":
		result = task.Arg1 / task.Arg2
	default:
		result = math.NaN()
	}
	return result
}

func sendResult(id string, result float64) error {
	// Send result to orchestrator
	reqBody := ResultRequest{ID: id, Result: result}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/internal/task/result", "application/json", string(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send result, status: %d", resp.StatusCode)
	}

	return nil
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
