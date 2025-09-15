package configure

import (
	"context"

	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ConfigCmd.Flags().BoolP("redo", "r", false, "repeat all configuration steps")
}

var ConfigCmd = &cobra.Command{
	Use:   "configure",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client := notion.NewClient()

		_, err := config.Load()
		if err != nil {
			return err
		}

		// Flags
		redo, err := cmd.Flags().GetBool("redo")
		if err != nil {
			return err
		}

		// Set task db
		tasksDatabaseID := config.TasksDatabaseID()
		if exit, err := ui.NewTextInput(
			"Tasks Database ID",
			&tasksDatabaseID,
			tasksDatabaseID,
		).Run(); err != nil || exit {
			return err
		}
		viper.Set(config.KeyTasksDatabaseID, tasksDatabaseID)

		// Set projects db
		projectsDatabaseID := config.ProjectsDatabaseID()
		if exit, err := ui.NewTextInput(
			"Projects Database ID",
			&projectsDatabaseID,
			projectsDatabaseID,
		).Run(); err != nil || exit {
			return err
		}
		viper.Set(config.KeyProjectsDatabaseID, projectsDatabaseID)

		// Set sprints db
		sprintsDatabaseID := config.SprintsDatabaseID()
		if exit, err := ui.NewTextInput(
			"Sprints Database ID",
			&sprintsDatabaseID,
			sprintsDatabaseID,
		).Run(); err != nil || exit {
			return err
		}
		viper.Set(config.KeySprintsDatabaseID, sprintsDatabaseID)

		// Emotes
		if !redo && !viper.IsSet(config.KeyUseEmotes) {
			useEmotes := config.UseEmotes()
			if exit, err := ui.NewSelectInput(
				"Do you want to use emotes (✅, 🔴, 🚀) in outputs?",
				[]ui.SelectItem[bool]{
					ui.NewSelectItem("Yes", true),
					ui.NewSelectItem("No", false),
				},
				&useEmotes,
			).Run(); err != nil || exit {
				return err
			}
			viper.Set(config.KeyUseEmotes, useEmotes)
		}

		// Fetch all users
		if _, err = ui.Spin(
			"Loading users",
			func() error {
				fetcherUsers := client.NewUserFetcher(ctx, true)
				users, err := fetcherUsers.All()
				if err != nil {
					return err
				}
				viper.Set(config.KeyUsers, users)
				return nil
			},
		); err != nil {
			return err
		}

		// Fetch all projects
		if _, err = ui.Spin(
			"Loading projects",
			func() error {
				fetcherProjects := client.NewProjectFetcher(ctx, config.ProjectsDatabaseID())
				projects, err := fetcherProjects.All()
				if err != nil {
					return err
				}
				viper.Set(config.KeyProjects, projects)
				return nil
			},
		); err != nil {
			return err
		}

		filename, err := config.Save()
		if err != nil {
			return err
		}
		ui.PrintlnfInfo("Configuration saved to %s", filename)

		return nil
	},
}
