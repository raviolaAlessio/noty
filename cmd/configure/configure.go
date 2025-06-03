package configure

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
}

func release(tp *tea.Program) {
	if err := tp.ReleaseTerminal(); err != nil {
		log.Fatal(err)
	}
}

var ConfigCmd = &cobra.Command{
	Use: "configure",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client := notion.NewClient()

		_, err := config.Load()
		if err != nil {
			return err
		}

		exit := false

		// Set task table
		viper.SetDefault("tasks_database_id", "")
		pTasksDatabaseID := config.TasksDatabaseID()
		tasksDatabaseID := pTasksDatabaseID
		tp := tea.NewProgram(ui.NewTextInput(
			"Tasks Database ID",
			&tasksDatabaseID,
			pTasksDatabaseID,
			&exit,
		))
		if _, err := tp.Run(); err != nil {
			return err
		}
		if exit {
			release(tp)
			return nil
		}
		viper.Set("tasks_database_id", tasksDatabaseID)

		// Set users table
		viper.SetDefault("projects_database_id", "")
		pProjectsDatabseID := config.ProjectsDatabaseID()
		projectsDatabaseID := pProjectsDatabseID
		tp = tea.NewProgram(ui.NewTextInput(
			"Projects Database ID",
			&projectsDatabaseID,
			pProjectsDatabseID,
			&exit,
		))
		if _, err := tp.Run(); err != nil {
			return err
		}
		if exit {
			release(tp)
			return nil
		}
		viper.Set("projects_database_id", projectsDatabaseID)

		// Fetch all users
		ui.PrintlnfInfo("Loading users...")
		fetcherUsers := client.NewUserFetcher(ctx, true)
		users, err := fetcherUsers.All()
		if err != nil {
			return err
		}
		viper.Set("users", users)
		ui.PrintlnfInfo("Users loaded")

		// Fetch all projects
		ui.PrintlnfInfo("Loading projects...")
		fetcherProjects := client.NewProjectFetcher(ctx, config.ProjectsDatabaseID())
		projects, err := fetcherProjects.All()
		if err != nil {
			return err
		}
		viper.Set("projects", projects)
		ui.PrintlnfInfo("Projects loaded")

		filename, err := config.Save()
		if err != nil {
			return err
		}
		ui.PrintlnfInfo("Configuration saved to %s", filename)

		return nil
	},
}
