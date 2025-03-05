package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Expression struct {
	Id          uuid.UUID `json:"id"`
	Status      string    `json:"status" binding:"oneof=active completed calculated"`
	Result      float64   `json:"result"`
	requestData string
	stack       []string
	tmpStack    []float64
}

type ExpressionsMap struct {
	m  map[uuid.UUID]Expression
	mu sync.RWMutex
	tm TaskMap
}

func NewExpressionsMap() *ExpressionsMap {
	return &ExpressionsMap{
		m:  make(map[uuid.UUID]Expression),
		tm: *NewTasksMap(),
	}
}

func (em *ExpressionsMap) GetExpression(id string) (*Expression, int) {
	// распарисм строку
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	em.mu.Lock()
	data, exists := em.m[uid]
	em.mu.Unlock()
	if !exists {
		return nil, http.StatusNotFound
	}
	return &data, http.StatusOK
}

func (em *ExpressionsMap) setTaskResult(taskBody TaskResultBody) int {
	em.tm.mu.Lock()
	defer em.tm.mu.Unlock()
	uid, err := uuid.Parse(taskBody.Id)
	if err != nil {
		return http.StatusInternalServerError
	}
	task, exists := em.tm.m[uid]
	if !exists {
		return http.StatusNotFound
	}
	delete(em.tm.m, task.Id)

	em.mu.Lock()
	defer em.mu.Unlock()

	expression, ex := em.m[task.expId]
	if !ex {
		return http.StatusInternalServerError
	}
	expression.Status = "active"
	if len(expression.stack) > 0 {
		if len(expression.tmpStack) > 0 {
			expression.tmpStack = append(expression.tmpStack[:1], append([]float64{taskBody.Result}, expression.tmpStack[1:]...)...)
		} else {
			expression.tmpStack = append(expression.tmpStack[:0], append([]float64{taskBody.Result}, expression.tmpStack[0:]...)...)
		}
	} else {
		expression.Result = taskBody.Result
		expression.Status = "completed"
	}

	em.m[expression.Id] = expression

	return http.StatusOK
}

func (em *ExpressionsMap) MoveTaskToStack(task Task) bool {
	em.tm.mu.Lock()
	defer em.tm.mu.Unlock()
	task, exists := em.tm.m[task.Id]
	if !exists {
		return false
	}
	delete(em.tm.m, task.Id)

	em.mu.Lock()
	defer em.mu.Unlock()
	expression, ex := em.m[task.expId]
	if !ex {
		return false
	}
	expression.Status = "active"

	expression.stack = append(expression.stack[:0], append([]string{task.Operation}, expression.stack[0:]...)...)
	expression.tmpStack = append(expression.tmpStack, task.Arg2)
	expression.tmpStack = append(expression.tmpStack, task.Arg1)

	em.m[expression.Id] = expression
	return true
}

func (em *ExpressionsMap) getTask() *Task {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, exp := range em.m {
		if exp.Status == "completed" || exp.Status == "calculated" {
			continue
		}
		for i, token := range exp.stack {
			if isMathOperator(token) {
				b := exp.tmpStack[len(exp.tmpStack)-1]
				a := exp.tmpStack[len(exp.tmpStack)-2]
				exp.tmpStack = exp.tmpStack[:len(exp.tmpStack)-2]

				for j := 0; j <= i; j++ {
					exp.stack = append(exp.stack[:0], exp.stack[1:]...)
				}

				exp.Status = "calculated"
				em.m[exp.Id] = exp

				em.tm.mu.Lock()
				defer em.tm.mu.Unlock()
				channel := make(chan Task)
				switch token {
				case "+":
					TIME_ADDITION_MS, exists := os.LookupEnv("TIME_ADDITION_MS")
					if !exists {
						TIME_ADDITION_MS = "1000"
					}
					TIME_ADDITION_MS_INT, err := time.ParseDuration(TIME_ADDITION_MS)
					if err != nil {
						TIME_ADDITION_MS_INT = 1000
					}
					task := Task{
						Id:            uuid.New(),
						expId:         exp.Id,
						Arg1:          a,
						Arg2:          b,
						Operation:     token,
						OperationTime: TIME_ADDITION_MS_INT,
					}
					em.tm.m[task.Id] = task

					go func() { channel <- task }()
					time.AfterFunc(TIME_ADDITION_MS_INT*time.Millisecond, func() {
						task := <-channel
						res := expMap.MoveTaskToStack(task)
						if res {
							fmt.Println("Время выполнения опрерации сложения истекло.")
						}
					})
					return &task
				case "-":
					TIME_SUBTRACTION_MS, exists := os.LookupEnv("TIME_SUBTRACTION_MS")
					if !exists {
						TIME_SUBTRACTION_MS = "1000"
					}
					TIME_SUBTRACTION_MS_INT, err := time.ParseDuration(TIME_SUBTRACTION_MS)
					if err != nil {
						TIME_SUBTRACTION_MS_INT = 1000
					}

					task := Task{
						Id:            uuid.New(),
						expId:         exp.Id,
						Arg1:          a,
						Arg2:          b,
						Operation:     token,
						OperationTime: TIME_SUBTRACTION_MS_INT,
					}
					em.tm.m[task.Id] = task
					go func() { channel <- task }()
					time.AfterFunc(TIME_SUBTRACTION_MS_INT*time.Millisecond, func() {
						task := <-channel
						res := expMap.MoveTaskToStack(task)
						if res {
							fmt.Println("Время выполнения опрерации вычитания истекло.")
						}
					})
					return &task
				case "*":
					TIME_MULTIPLICATIONS_MS, exists := os.LookupEnv("TIME_MULTIPLICATIONS_MS")
					if !exists {
						TIME_MULTIPLICATIONS_MS = "1000"
					}
					TIME_MULTIPLICATIONS_MS_INT, err := time.ParseDuration(TIME_MULTIPLICATIONS_MS)
					if err != nil {
						TIME_MULTIPLICATIONS_MS_INT = 1000
					}
					task := Task{
						Id:            uuid.New(),
						expId:         exp.Id,
						Arg1:          a,
						Arg2:          b,
						Operation:     token,
						OperationTime: TIME_MULTIPLICATIONS_MS_INT,
					}
					em.tm.m[task.Id] = task
					go func() { channel <- task }()
					time.AfterFunc(TIME_MULTIPLICATIONS_MS_INT*time.Millisecond, func() {
						task := <-channel
						res := expMap.MoveTaskToStack(task)
						if res {
							fmt.Println("Время выполнения опрерации умножения истекло.")
						}
					})
					return &task
				case "/":
					TIME_DIVISIONS_MS, exists := os.LookupEnv("TIME_DIVISIONS_MS")
					if !exists {
						TIME_DIVISIONS_MS = "1000"
					}
					TIME_DIVISIONS_MS_INT, err := time.ParseDuration(TIME_DIVISIONS_MS)
					if err != nil {
						TIME_DIVISIONS_MS_INT = 1000
					}
					task := Task{
						Id:            uuid.New(),
						expId:         exp.Id,
						Arg1:          a,
						Arg2:          b,
						Operation:     token,
						OperationTime: TIME_DIVISIONS_MS_INT,
					}
					em.tm.m[task.Id] = task
					go func() { channel <- task }()
					time.AfterFunc(TIME_DIVISIONS_MS_INT*time.Millisecond, func() {
						task := <-channel
						res := expMap.MoveTaskToStack(task)
						if res {
							fmt.Println("Время выполнения опрерации деления истекло.")
						}
					})
					return &task
				default:
					return nil
				}
			} else {
				value, _ := strconv.ParseFloat(token, 64)
				exp.tmpStack = append(exp.tmpStack, value)
			}
		}
	}
	return nil
}

// Добавление выражения для вычисления
func (em *ExpressionsMap) AddExpression(expression string) (int, *Expression) {

	allstack, success := createStack(expression)

	if !success {
		return http.StatusUnprocessableEntity, nil
	}

	exp := Expression{
		Id:          uuid.New(),
		Status:      "active",
		requestData: expression,
		stack:       allstack,
		tmpStack:    make([]float64, 0),
	}

	em.mu.Lock()
	em.m[exp.Id] = exp
	em.mu.Unlock()
	return http.StatusCreated, &exp
}

// Получение всех выражений для вычисления
func (em *ExpressionsMap) GetExpressions() []Expression {
	expressions := make([]Expression, 0, 1)
	em.mu.Lock()
	for _, e := range em.m {
		expressions = append(expressions, e)
	}
	em.mu.Unlock()
	return expressions
}
