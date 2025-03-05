package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type TaskBody struct {
	Task Task `json:"task"`
}

type Task struct {
	Id            uuid.UUID     `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

type TaskResult struct {
	Id     string  `json:"id"`
	Result float64 `json:"result"`
}

var sleepTime = 1 * time.Second

func calculate() {
	client := &http.Client{}
	for {
		resp, err := http.Get("http://localhost:8000/api/internal/task")
		if err != nil {
			fmt.Println("Оркестратор не запущен")
			time.Sleep(sleepTime)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Нет задач")
			time.Sleep(sleepTime)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Println(err)
			time.Sleep(sleepTime)
			continue
		}

		var taskBody TaskBody

		errP := json.Unmarshal([]byte(body), &taskBody)
		if errP != nil {
			fmt.Println(errP)
			time.Sleep(sleepTime)
			continue
		}

		task := taskBody.Task

		var result float64
		switch task.Operation {
		case "+":
			result = task.Arg1 + task.Arg2
		case "-":
			result = task.Arg1 - task.Arg2
		case "*":
			result = task.Arg1 * task.Arg2
		case "/":
			if task.Arg2 == 0 {
				result = 0
			}
			result = task.Arg1 / task.Arg2
		default:
			result = 0
		}

		taskRes := &TaskResult{
			Id:     task.Id.String(),
			Result: result,
		}

		data, err := json.Marshal(taskRes)
		if err != nil {
			fmt.Println(err)
			time.Sleep(sleepTime)
			continue
		}
		// Создаем новый запрос
		req, err := http.NewRequest("POST", "http://localhost:8000/api/internal/task", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println(err)
			time.Sleep(sleepTime)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		respRes, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			time.Sleep(sleepTime)
			continue
		}

		fmt.Println("Статус установки результата в оркестратор: ", respRes.Status)
	}
}

func main() {
	COMPUTING_POWER, exists := os.LookupEnv("COMPUTING_POWER")
	if !exists {
		COMPUTING_POWER = "2"
	}

	computing_power_int, err := strconv.Atoi(COMPUTING_POWER)

	if err != nil {
		computing_power_int = 2
	}
	for i := 0; i < computing_power_int; i++ {
		go calculate()
	}

	select {}
}
