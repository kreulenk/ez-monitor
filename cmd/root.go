package cmd

import (
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"github.com/kreulenk/ez-monitor/pkg/tui"
	"github.com/spf13/cobra"
	"os"
)

func genRootCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "ez-monitor <inventory-file>",
		Short: "An agentless SSH based system monitoring tool",
		Long:  `EZ-Monitor allows you to easily monitor your Linux infrastructure by only requiring SSH connections`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			filename := args[0]
			inventoryInfo, err := inventory.LoadInventory(filename)
			cobra.CheckErr(err)

			statsChan, err := statistics.StartStatisticsCollection(ctx, inventoryInfo)
			cobra.CheckErr(err)
			tui.Initialize(ctx, inventoryInfo, statsChan)
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
