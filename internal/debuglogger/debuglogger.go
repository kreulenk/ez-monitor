// The debuglogger package is used during development to allow for the easy addition
// of slog.Info() logs to be printed to a ez-monitor.log file

package debuglogger

import (
	"fmt"
	"log/slog"
	"os"
)

func Initialize() *os.File {
	if _, ok := os.LookupEnv("EZ_MONITOR_DEBUG"); ok {
		file, err := os.Create("ez-monitor.log")
		if err != nil {
			fmt.Println("unable to create debug log file")
		}
		handler := slog.NewTextHandler(file, nil)
		logger := slog.New(handler)
		slog.SetDefault(logger)
		return file
	} else {
		file, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
		handler := slog.NewTextHandler(file, nil)
		logger := slog.New(handler)
		slog.SetDefault(logger)
		return file
	}
}
