package task

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
)

// choiceValue implements the [pflag.Value] interface.
type choiceValue struct {
	value    string
	validate func(string) error
}

// Set sets the value of the choice.
func (f *choiceValue) Set(s string) error {
	err := f.validate(s)
	if err != nil {
		return err
	}

	f.value = s
	return nil
}

// Type returns the type of the choice, which must be "string" for [pflag.FlagSet.GetString].
func (f *choiceValue) Type() string { return "string" }

// String returns the current value of the choice.
func (f *choiceValue) String() string { return f.value }

// StringChoice returns a [choiceValue] that validates the value against a set
// of choices. Only the last value will be used if multiple values are set.
func StringChoice(choices []string) *choiceValue {
	return &choiceValue{
		validate: func(s string) error {
			if slices.Contains(choices, s) {
				return nil
			}
			return fmt.Errorf("must be one of %v", choices)
		},
	}
}

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
	TaskCmd.Flags().StringSliceP("status", "s", []string{}, "filter tasks by status(es) [NS, P, TBT, T, D, ND]")

	// Sprint
	TaskCmd.Flags().Var(StringChoice([]string{
		"default", "all", "backlog", "current",
	}), "sprint", "sprint to search taskts in")

	// Limits
	TaskCmd.Flags().Bool("all", false, "fetch all tasks")
	TaskCmd.Flags().IntP("limit", "l", 50, "limit the number of tasks to fetch")
	TaskCmd.MarkFlagsMutuallyExclusive("all", "limit")

	// Export
	TaskCmd.Flags().StringP("outfile", "o", "", "export result as csv")
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
				case "P":
					filter.Statuses = append(filter.Statuses, notion.StatusInProgress)
				case "TBT":
					filter.Statuses = append(filter.Statuses, notion.StatusToBeTested)
				case "T":
					filter.Statuses = append(filter.Statuses, notion.StatusInTesting)
				case "D":
					filter.Statuses = append(filter.Statuses, notion.StatusDone)
				case "ND":
					filter.Statuses = append(filter.Statuses, notion.StatusNotDone)
				default:
					return fmt.Errorf("unknown status '%s', valid values are [NS, P, TBT, T, D]", status)
				}
			}
		}

		// Sprint Flag
		if sprint, err := cmd.Flags().GetString("sprint"); err != nil {
			return err
		} else if sprint == "default" {
			filter.Sprint = notion.TaskSprintNoBacklog{}
		} else if sprint == "all" {
			filter.Sprint = nil
		} else if sprint == "backlog" {
			filter.Sprint = notion.TaskSprintOnlyBacklog{}
		} else if sprint == "current" {
			// Fetch sprint
			s := "Current"
			sprintFetcher := notionClient.NewSprintFetcher(
				ctx,
				config.SprintsDatabaseID(),
				notion.SprintFilter{
					Status: &s,
				},
			)
			res, err := sprintFetcher.NextOne()
			if err != nil {
				return err
			}
			// Set ID as filter
			filter.Sprint = notion.TaskSprintByID{
				ID: res.ID,
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
			keyStoryId     = "storyId"
			keyProject     = "project"
			keyName        = "name"
			keyStatus      = "status"
			keyAssignee    = "assignee"
			keyReviewer    = "reviewer"
			keyPriority    = "priority"
			keyEstimate    = "estimate"
			keyCreatedTime = "created_time"
		)
		columns := []ui.TableColumn{
			ui.NewTableColumn(keyId, "ID", false).WithAlignment(ui.TableRight),
			ui.NewTableColumn(keyStoryId, "Story ID", true),
			ui.NewTableColumn(keyProject, "Project", true),
			ui.NewTableColumn(keyName, "Name", true),
			ui.NewTableColumn(keyAssignee, "Assignee", true),
			ui.NewTableColumn(keyReviewer, "Reviewer", true),
			ui.NewTableColumn(keyStatus, "Status", true).WithValueFunc(
				func(value string) string {
					if config.UseEmotes() {
						emote := config.StatusEmote(value)
						if emote != "" {
							value = fmt.Sprintf("%s %s", emote, value)
						}
					}
					return value
				},
			),
			ui.NewTableColumn(keyEstimate, "Estimate", true).WithAlignment(ui.TableRight),
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

		timeFormat := config.DatetimeFormat()

		// Add rows
		rows := make([]ui.TableRow, 0, len(tasks))
		for _, task := range tasks {
			project := ""
			if (len(task.ProjectID) > 0) {
				project = projectsMap[task.ProjectID[0]]
			}
			rows = append(rows, ui.TableRow{
				keyId:          task.ID,
				keyStoryId:     task.StoryID,
				keyProject:     project,
				keyName:        task.Name,
				keyAssignee:    strings.Join(task.Assignee, ", "),
				keyReviewer:    strings.Join(task.Reviewer, ", "),
				keyStatus:      task.Status,
				keyEstimate:    fmt.Sprintf("%.1f h", task.Estimate),
				keyPriority:    task.Priority,
				keyCreatedTime: task.Created.Local().Format(timeFormat),
			})
		}
		// Render result
		table := ui.NewTable(columns).WithRows(rows)
		fmt.Println(table.Render())

		resultLog := fmt.Sprintf("\nFetched %d tasks", len(rows))
		if taskFetcher.HasMore() {
			resultLog += ", has more"
		}
		ui.PrintlnInfo(resultLog)

		// Export
		if outfile, err := cmd.Flags().GetString("outfile"); err != nil {
			return err
		} else if outfile != "" {
			abs, err := filepath.Abs(outfile)
			if err != nil {
				return err
			}
			err = table.ExportCSV(abs)
			if err != nil {
				ui.PrintlnfWarn("Could not export to CSV: %s", err.Error())
			} else {
				ui.PrintlnfInfo("Data exported to CSV file %s", abs)
			}
		}

		return nil
	},
}
