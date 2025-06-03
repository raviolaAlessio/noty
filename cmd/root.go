package cmd

import (
	"fmt"
	"os"

	"github.com/ravvio/noty/cmd/configure"
	"github.com/ravvio/noty/cmd/task"
	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/ui"
	"github.com/spf13/cobra"
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		ui.PrintlnfError("%s", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(configure.ConfigCmd)
	rootCmd.AddCommand(task.TaskCmd)
}

var rootCmd = &cobra.Command{
	Use:   "noty",
	Short: "A utility to manage notion tasks",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := config.Init()
		if err != nil {
			return err
		}

		if cmd.CalledAs() == "config" {
			fmt.Println("config cmd")
			return nil
		}

		ok, err := config.Load()
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("configuration not found, run the 'configure' command to generate it")
		}
		return nil
	},
}
