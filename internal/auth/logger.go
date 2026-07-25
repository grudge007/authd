package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	TaskID    string    `json:"task_id"`
	Event     string    `json:"event"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	errorLogger   *log.Logger
	successLogger *log.Logger
	mu            sync.Mutex
)

func init() {
	// Create logs directory if it doesn't exist
	logsDir := filepath.Join("logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		return
	}

	// Create error log file
	errorFile, err := os.OpenFile(filepath.Join(logsDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open error log file: %v\n", err)
		return
	}
	errorLogger = log.New(errorFile, "", 0)

	// Create success log file
	successFile, err := os.OpenFile(filepath.Join(logsDir, "success.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open success log file: %v\n", err)
		return
	}
	successLogger = log.New(successFile, "", 0)
}

func GenerateTaskID() string {
	return uuid.New().String()
}

func LogError(event, message string, err error) {
	mu.Lock()
	defer mu.Unlock()

	taskID := GenerateTaskID()
	entry := LogEntry{
		TaskID:    taskID,
		Event:     event,
		Status:    "error",
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		entry.Message = fmt.Sprintf("%s: %v", message, err)
	}

	jsonBytes, _ := json.Marshal(entry)
	errorLogger.Println(string(jsonBytes))
}

func LogSuccess(event, message string) {
	mu.Lock()
	defer mu.Unlock()

	taskID := GenerateTaskID()
	entry := LogEntry{
		TaskID:    taskID,
		Event:     event,
		Status:    "success",
		Message:   message,
		Timestamp: time.Now(),
	}

	jsonBytes, _ := json.Marshal(entry)
	successLogger.Println(string(jsonBytes))
}
