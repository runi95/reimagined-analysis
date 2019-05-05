package logger

import (
	"fmt"
	"github.com/logrusorgru/aurora"
)

type Logger struct {}

func (*Logger) Info(message string) {
	fmt.Printf("[%s] - %s\n", aurora.Green("INFO"), message)
}

func (*Logger) Warning(message string) {
	fmt.Printf("[%s] - %s\n", aurora.Yellow("Warning"), message)
}

func (*Logger) Error(message string) {
	fmt.Printf("[%s] - %s\n", aurora.Red("Error"), message)
}