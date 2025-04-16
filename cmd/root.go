package cmd

import (
	"ez-monitor/pkg/inventory"
	"ez-monitor/pkg/statistics"
	"ez-monitor/pkg/tui"
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"text/tabwriter"
)

func genRootCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "ez-monitor <inventory-file>",
		Short: "An agentless SSH based system monitoring tool",
		Long:  `ez-monitor allows you to easily monitor your infrastructure by only requiring SSH connections`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			filename := args[0]
			inventoryInfo, err := inventory.LoadInventory(filename)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			statsChan, err := statistics.StartStatisticsCollection(ctx, inventoryInfo)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			tui.Initialize(ctx, inventoryInfo, statsChan)

			//printStats(statsChan)
		},
	}
	return cmd
}

func Execute() {
	rootCmd := genRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// TODO remove this function once the TUI is fully in place
func printStats(statsChan chan statistics.HostStats) {
	mostRecentStat := map[string]*statistics.HostStats{}

	// Create a new tab writer
	for s := range statsChan {
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
		mostRecentStat[s.NameOfHost] = &s

		fmt.Fprintln(writer, "Name\tCPU usage (%)\tMem Usage\tMem Total\tMem Percentage")
		// Write the statistics
		for _, statistic := range mostRecentStat {
			if statistic.Error == nil {
				memoryPercentage := (statistic.MemoryUsage / statistic.MemoryTotal) * 100
				fmt.Fprintf(writer, "%s\t%.2f\t%.2f\t%.2f\t%.2f\n",
					statistic.NameOfHost,
					statistic.CPUUsage,
					statistic.MemoryUsage,
					statistic.MemoryTotal,
					memoryPercentage,
				)
			}
		}

		// Flush the writer to ensure all output is written
		writer.Flush()

		for _, statistic := range mostRecentStat {
			if statistic.Error != nil {
				fmt.Printf("failed to connect to stat %s: %s\n", statistic.NameOfHost, statistic.Error)
			}
		}
	}
}
