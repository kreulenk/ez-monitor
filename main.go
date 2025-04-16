package main

import (
	"github.com/kreulenk/ez-monitor/cmd"
	"github.com/kreulenk/ez-monitor/internal/debuglogger"
)

func main() {
	logFile := debuglogger.Initialize()
	defer logFile.Close()
	cmd.Execute()
}
