package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
)

func init() {
	// Users
	TaskCmd.Flags().StringP("user", "u", "", "filter tasks by user (assignee or reviewer)")
	TaskCmd.Flags().StringP("assignee", "a", "", "filter tasks by assignee")
	TaskCmd.Flags().StringP("reviewer", "r", "", "filter tasks by reviewer")
	TaskCmd.MarkFlagsMutuallyExclusive("assignee", "user")
	TaskCmd.MarkFlagsMutuallyExclusive("reviewer", "user")

	// Project
	TaskCmd.Flags().StringSliceP("project", "p", []string{}, "filter by project(s)")

	// Status
	TaskCmd.Flags().StringSliceP("status", "s", []string{}, "filter tasks by status(es) [NS, P, TBT, T, D]")

	// Limits
	TaskCmd.Flags().Bool("all", false, "fetch all tasks")
	TaskCmd.Flags().IntP("limit", "l", 50, "limit the number of tasks to fetch")
	TaskCmd.MarkFlagsMutuallyExclusive("all", "limit")
}

var TaskCmd = &cobra.Command{
	Use:   "task",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		notionClient := notion.NewClient()

		// Load config
		usersList := config.Users()
		projectsList := config.Projects()
		projectsMap := config.ProjectsMap()

		// Create filter
		filter := notion.TaskFilter{}

		// Assignee Flag
		if assignee, err := cmd.Flags().GetString("assignee"); err != nil {
			return err
		} else if assignee != "" {
			users := config.Users()
			for _, user := range users {
				if strings.Contains(strings.ToLower(user.Name), strings.ToLower(assignee)) {
					filter.Assignee = &user.ID
					break
				}
			}
			if filter.Assignee == nil {
				return fmt.Errorf("no user found for assignee '%s'", assignee)
			}
		}
		// Reviewer Flag
		if reviewer, err := cmd.Flags().GetString("reviewer"); err != nil {
			return err
		} else if reviewer != "" {
			for _, user := range usersList {
				if strings.Contains(strings.ToLower(user.Name), strings.ToLower(reviewer)) {
					filter.Reviewer = &user.ID
					break
				}
			}
			if filter.Reviewer == nil {
				return fmt.Errorf("no user found for reviewer '%s'", reviewer)
			}
		}
		// User Flag
		if username, err := cmd.Flags().GetString("user"); err != nil {
			return err
		} else if username != "" {
			for _, user := range usersList {
				if strings.Contains(strings.ToLower(user.Name), strings.ToLower(username)) {
					filter.User = &user.ID
					break
				}
			}
			if filter.User == nil {
				return fmt.Errorf("no user found for '%s'", username)
			}
		}
		// Projects Flag
		if projects, err := cmd.Flags().GetStringSlice("project"); err != nil {
			return err
		} else if len(projects) > 0 {
			for _, projectName := range projects {
				l := len(filter.Projects)

				projectName = strings.ToLower(projectName)
				for _, project := range projectsList {
					if strings.Contains(
						strings.ToLower(project.Name),
						projectName,
					) {
						filter.Projects = append(filter.Projects, project.ID)
					}
				}

				if l == len(filter.Projects) {
					return fmt.Errorf("no project found for '%s'", projectName)
				}
			}
		}
		// Status Flag
		statuses, err := cmd.Flags().GetStringSlice("status")
		if err != nil {
			return err
		}
		if len(statuses) > 0 {
			for _, status := range statuses {
				switch strings.ToUpper(status) {
				case "NS":
					filter.Statuses = append(filter.Statuses, notion.StatusNotStarted)
					break
				case "P":
					filter.Statuses = append(filter.Statuses, notion.StatusInProgress)
					break
				case "TBT":
					filter.Statuses = append(filter.Statuses, notion.StatusToBeTested)
					break
				case "T":
					filter.Statuses = append(filter.Statuses, notion.StatusInTesting)
					break
				case "D":
					filter.Statuses = append(filter.Statuses, notion.StatusDone)
					break
				default:
					return fmt.Errorf("unknown status '%s', valid values are [NS, P, TBT, T, D]", status)
				}
			}
		}

		// Create fetcher
		taskFetcher := notionClient.NewTaskFetcher(
			ctx,
			config.TasksDatabaseID(),
			filter,
		)

		// All / Limit Flag
		if all, err := cmd.Flags().GetBool("all"); err != nil {
			return fmt.Errorf("failed request: %s", err)
		} else if all == false {
			if limit, err := cmd.Flags().GetInt("limit"); err != nil {
				return err
			} else {
				taskFetcher = taskFetcher.WithLimit(limit)
			}
		}

		// Fetch
		tasks, err := taskFetcher.All()
		if err != nil {
			return err
		}

		// Setup table
		var (
			keyId          = "id"
			keyProject     = "project"
			keyName        = "name"
			keyStatus      = "status"
			keyAssignee    = "assignee"
			keyReviewer    = "reviewer"
			keyPriority    = "priority"
			keyCreatedTime = "created_time"
		)
		columns := []ui.TableColumn{
			ui.NewTableColumn(keyId, "ID", false).WithAlignment(ui.TableRight),
			ui.NewTableColumn(keyProject, "Project", true),
			ui.NewTableColumn(keyName, "Name", true),
			ui.NewTableColumn(keyAssignee, "Assignee", true),
			ui.NewTableColumn(keyReviewer, "Reviewer", true),
			ui.NewTableColumn(keyStatus, "Status", true).WithValueFunc(
				func(value string) string {
					if !config.UseEmotes() {
						return value
					}

					emote := config.StatusEmote(value)
					if emote != "" {
						value = fmt.Sprintf("%s %s", emote, value)
					}
					return value
				},
			),
			ui.NewTableColumn(keyPriority, "Priority", true).WithStyleFunc(
				func(style lipgloss.Style, value string) lipgloss.Style {
					switch value {
					case "High":
						return style.Foreground(ui.PriorityHigh)
					case "Medium":
						return style.Foreground(ui.PriorityMedium)
					case "Low":
						return style.Foreground(ui.PriorityLow)
					}
					return style
				},
			),
			ui.NewTableColumn(keyCreatedTime, "Created", true),
		}

		// Add rows
		rows := make([]ui.TableRow, 0, len(tasks))
		for _, task := range tasks {
			rows = append(rows, ui.TableRow{
				keyId:          task.ID,
				keyProject:     projectsMap[task.ProjectID[0]],
				keyName:        task.Name,
				keyAssignee:    strings.Join(task.Assignee, ", "),
				keyReviewer:    strings.Join(task.Reviewer, ", "),
				keyStatus:      task.Status,
				keyPriority:    task.Priority,
				keyCreatedTime: task.Created.Format(time.RFC3339),
			})
		}
		// Render result
		table := ui.NewTable(columns).WithRows(rows)
		fmt.Println(table.Render())

		resultLog := fmt.Sprintf("\nFetched %d tasks", len(rows))
		if taskFetcher.HasMore() {
			resultLog += ", has more"
		}
		ui.PrintlnfInfo(resultLog)

		return nil
	},
}
