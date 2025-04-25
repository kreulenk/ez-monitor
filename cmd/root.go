package cmd

import (
	"errors"
	"fmt"
	"github.com/kreulenk/ez-monitor/internal/build"
	"github.com/kreulenk/ez-monitor/pkg/inventory"
	"github.com/kreulenk/ez-monitor/pkg/statistics"
	"github.com/kreulenk/ez-monitor/pkg/tui"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

func genRootCmd() *cobra.Command {
	var version bool
	var hostToAddEncryptedPassword string

	var cmd = &cobra.Command{
		Use:   "ez-monitor <inventory-file>",
		Short: "An agentless SSH based system monitoring tool",
		Long:  `EZ-Monitor allows you to easily monitor your Linux infrastructure by only requiring SSH connections`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 && !version {
				return errors.New("no inventory file was provided")
			}
			if len(args) > 1 {
				return errors.New("too many arguments were provided. You must specify a single inventory file")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			if version {
				fmt.Printf("%s\n%s_%s\n%s\n", build.Version, runtime.GOOS, runtime.GOARCH, build.SHA)
				return
			}
			if hostToAddEncryptedPassword != "" {
				err := inventory.BeginPasswordEncryptFlow(hostToAddEncryptedPassword, args[0])
				cobra.CheckErr(err)
				os.Exit(0)
			}

			inventoryInfo, err := inventory.LoadInventory(args[0])
			cobra.CheckErr(err)

			statsChan, err := statistics.StartStatisticsCollection(ctx, inventoryInfo)
			cobra.CheckErr(err)
			tui.Initialize(ctx, inventoryInfo, statsChan)
		},
	}
	cmd.Flags().BoolVarP(&version, "version", "v", false, "Show version information")
	cmd.Flags().StringVar(&hostToAddEncryptedPassword, "add-encrypted-pass", "", "Specify the alias of a host in the host file for which you'd like to set an encrypted password")

	return cmd
}

func Execute() {
	rootCmd := genRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
