package main

import (
	"ez-monitor/cmd"
	"ez-monitor/internal/debuglogger"
)

func main() {
	logFile := debuglogger.Initialize()
	defer logFile.Close()
	cmd.Execute()
}
