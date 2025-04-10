package main

import (
	"ez-monitor/pkg/inventory"
	"ez-monitor/pkg/stats"
	"fmt"
	"os"
	"text/tabwriter"
)

func main() {
	filename := "./test/example-inventories/basic.ini"
	hosts, err := inventory.LoadInventory(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	statistics := stats.CollectHostStats(hosts)
	printStats(statistics)
}

func printStats(statistics []stats.HostStats) {
	// Create a new tab writer
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)

	// Write the header
	fmt.Fprintln(writer, "Name\tCPU usage (%)\tMem Usage\tMem Total\tMem Percentage")
	// Write the statistics
	for _, statistic := range statistics {
		if statistic.Error == nil {
			memoryPercentage := (statistic.MemoryUsage / statistic.MemoryTotal) * 100
			fmt.Fprintf(writer, "%s\t%.2f\t%.2f\t%.2f\t%.2f\n",
				statistic.Name,
				statistic.CPUUsage,
				statistic.MemoryUsage,
				statistic.MemoryTotal,
				memoryPercentage,
			)
		}
	}

	// Flush the writer to ensure all output is written
	writer.Flush()

	for _, statistic := range statistics {
		if statistic.Error != nil {
			fmt.Printf("failed to connect to host %s: %s\n", statistic.Name, statistic.Error)
		}
	}

}
