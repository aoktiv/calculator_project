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

var computingPower = 2

type ResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

func main() {
	computingPower = getEnv("COMPUTING_POWER", 2)

	for i := 0; i < computingPower; i++ {
		go taskWorker()
	}

	select {}
}

func taskWorker() {
	for {
		task, err := fetchTask()
		if err != nil {
			log.Println("Error fetching task:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		result := compute(task)

		err = sendResult(task.ID, result)
		if err != nil {
			log.Println("Error sending result:", err)
		}

		time.Sleep(1 * time.Second)
	}
}

func fetchTask() (orchestrator.Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return orchestrator.Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return orchestrator.Task{}, fmt.Errorf("failed to get task, status: %d", resp.StatusCode)
	}

	var result struct {
		Task orchestrator.Task `json:"task"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return orchestrator.Task{}, err
	}

	return result.Task, nil
}

func compute(task orchestrator.Task) float64 {
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
