package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Id            uuid.UUID `json:"id"`
	expId         uuid.UUID
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

type TaskMap struct {
	m  map[uuid.UUID]Task
	mu sync.RWMutex
}

func NewTasksMap() *TaskMap {
	return &TaskMap{
		m: make(map[uuid.UUID]Task),
	}
}
