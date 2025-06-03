package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
)

func init() {
	TaskCmd.Flags().StringP("assignee", "a", "", "filter tasks by assignee")
	TaskCmd.Flags().StringSliceP("status", "s", []string{}, "filter tasks by status(es) [NS, P, TBT, T, D]")
}

var TaskCmd = &cobra.Command{
	Use:   "task",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		notionClient := notion.NewClient()

		filter := notion.TaskFilter{}

		// Assignee Flag
		assignee, err := cmd.Flags().GetString("assignee")
		if err != nil {
			return err
		}
		if assignee != "" {
			users := config.Users()
			for _, user := range users {
				if strings.Contains(user.Name, assignee) {
					filter.Assignee = &user.ID
					break
				}
			}
			if filter.Assignee == nil {
				return fmt.Errorf("no user found for assignee %s", assignee)
			}
		}
		// Status Flag
		statuses, err := cmd.Flags().GetStringSlice("status")
		if err != nil {
			return err
		}
		if len(statuses) > 0 {
			var s []string
			for _, status := range statuses {
				switch status {
				case "NS":
					s = append(s, notion.StatusNotStarted)
					break
				case "P":
					s = append(s, notion.StatusInProgress)
					break
				case "TBT":
					s = append(s, notion.StatusToBeTested)
					break
				case "T":
					s = append(s, notion.StatusInTesting)
					break
				case "D":
					s = append(s, notion.StatusDone)
					break
				default:
					s = append(s, status)
					break
				}
			}
			filter.Statuses = s
		}

		// Fetch tasks
		taskFetcher := notionClient.NewTaskFetcher(
			ctx,
			config.TasksDatabaseID(),
			filter,
		).WithLimit(10)

		tasks, err := taskFetcher.All()
		if err != nil {
			return err
		}

		// Setup table
		var (
			keyId       = "id"
			keyProject  = "project"
			keyName     = "name"
			keyStatus   = "status"
			keyAssignee = "assignee"
			keyReviewer = "reviewer"
			keyPriority = "priority"
		)
		columns := []ui.TableColumn{
			ui.NewTableColumn(keyId, "ID", false).WithAlignment(ui.TableRight),
			ui.NewTableColumn(keyProject, "Project", true),
			ui.NewTableColumn(keyName, "Name", true),
			ui.NewTableColumn(keyAssignee, "Assignee", true),
			ui.NewTableColumn(keyReviewer, "Reviewer", true),
			ui.NewTableColumn(keyStatus, "Status", true),
			ui.NewTableColumn(keyPriority, "Priority", true),
		}

		projects := config.ProjectsMap()

		// Add rows
		rows := make([]ui.TableRow, 0, len(tasks))
		for _, task := range tasks {
			rows = append(rows, ui.TableRow{
				keyId:       task.ID,
				keyProject:  projects[task.ProjectID[0]],
				keyName:     task.Name,
				keyAssignee: strings.Join(task.Assignee, ", "),
				keyReviewer: strings.Join(task.Reviewer, ", "),
				keyStatus:   task.Status,
				keyPriority: task.Priority,
			})
		}
		// Render result
		table := ui.NewTable(columns).WithRows(rows)
		fmt.Println(table.Render())

		return nil
	},
}
