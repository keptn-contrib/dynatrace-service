package keptn

import (
	"encoding/json"
	"fmt"
)

type keptnLogMessage struct {
	KeptnContext string `json:"keptnContext"`
	Message      string `json:"message"`
	KeptnService string `json:"keptnService"`
	LogLevel     string `json:"logLevel"`
}

// Info logs an info message
func Info(keptnContext string, message string) {
	var logMessage keptnLogMessage
	logMessage.KeptnContext = keptnContext
	logMessage.LogLevel = "INFO"
	logMessage.KeptnService = "dynatrace-service"
	logMessage.Message = message

	printLogMessage(logMessage)
}

// Error logs an error message
func Error(keptnContext string, message string) {
	var logMessage keptnLogMessage
	logMessage.KeptnContext = keptnContext
	logMessage.LogLevel = "ERROR"
	logMessage.KeptnService = "dynatrace-service"
	logMessage.Message = message

	printLogMessage(logMessage)
}

// Debug logs a debug message
func Debug(keptnContext string, message string) {
	var logMessage keptnLogMessage
	logMessage.KeptnContext = keptnContext
	logMessage.LogLevel = "DEBUG"
	logMessage.KeptnService = "dynatrace-service"
	logMessage.Message = message

	printLogMessage(logMessage)
}

func printLogMessage(logMessage keptnLogMessage) {
	logString, err := json.Marshal(logMessage)

	if err != nil {
		fmt.Println("Could not log keptn log message")
		return
	}

	fmt.Println(string(logString))
}
