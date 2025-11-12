package cmd

import (
	"fmt"
	"os"

	"github.com/ravvio/noty/cmd/configure"
	"github.com/ravvio/noty/cmd/hours"
	"github.com/ravvio/noty/cmd/task"
	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/flags"
	"github.com/ravvio/noty/ui"
	"github.com/spf13/cobra"
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		ui.PrintlnfError("Error: %s", err)
		os.Exit(1)
	}
}

func init() {
	err := config.Init()
	if err != nil {
		ui.PrintlnfError("%s", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(configure.ConfigCmd)
	rootCmd.AddCommand(task.TaskCmd)
	rootCmd.AddCommand(hours.HoursCmd)

	rootCmd.PersistentFlags().Var(
		flags.StringChoice(
			[]string{"default", "md"},
			"default",
		),
		"style",
		"output table style [default, md]",
	)
}

var rootCmd = &cobra.Command{
	Use:           "noty",
	Short:         "A utility to manage notion tasks",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == configure.ConfigCmd.Use {
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
