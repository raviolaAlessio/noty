package configure

import (
	"context"
	"fmt"
	"strings"

	"github.com/ravvio/easycli-ui/espinner"
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

		// Set hours db
		hoursDatabaseID := config.HoursDatabaseID()
		if exit, err := ui.NewTextInput(
			"Hours Entries Database ID",
			&hoursDatabaseID,
			hoursDatabaseID,
		).Run(); err != nil || exit {
			return err
		}
		viper.Set(config.KeyHoursDatabaseID, hoursDatabaseID)

		// Emotes
		if redo || !viper.IsSet(config.KeyUseEmotes) {
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

		// Date format
		if redo || !viper.IsSet(config.KeyDateFormat) {
			dateFormat := config.DateFormat()
			if exit, err := ui.NewSelectInput(
				"Select your preferred date format",
				[]ui.SelectItem[string]{
					ui.NewSelectItem("2006-01-02 (Year-Month-Day)", "2006-01-02"),
					ui.NewSelectItem("02/01/2006 (Day/Month/Year)", "02/01/2006"),
				},
				&dateFormat,
			).Run(); err != nil || exit {
				return err
			}
			viper.Set(config.KeyDateFormat, dateFormat)
		}

		// Time format
		if redo || !viper.IsSet(config.KeyDatetimeFormat) {
			dateFormat := config.DateFormat()
			datetimeFormatSplit := strings.Split(config.DatetimeFormat(), " ")
			var timeFormat string
			if len(datetimeFormatSplit) < 2 {
				timeFormat = ""
			} else {
				timeFormat = datetimeFormatSplit[1]
			}

			if exit, err := ui.NewSelectInput(
				"Select your preferred time format",
				[]ui.SelectItem[string]{
					ui.NewSelectItem("15:04", "15:04"),
					ui.NewSelectItem("03:04 PM", "3:4 PM"),
				},
				&timeFormat,
			).Run(); err != nil || exit {
				return err
			}
			viper.Set(config.KeyDatetimeFormat, fmt.Sprintf("%s %s", dateFormat, timeFormat))
		}

		// Fetch all users
		s := espinner.NewSpinner(
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
		)
		if err = s.Spin(); err != nil {
			return err
		}

		// Fetch all projects
		espinner.NewSpinner(
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
		)
		if err = s.Spin(); err != nil {
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
