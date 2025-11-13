package task

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/flags"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
	"github.com/ravvio/noty/utils"
)

type TaskGroupingValues struct {
	Count int
	Hours float64
}

// Table column names
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
	keyCreatedTime = "createdTime"
	keyStoryURL    = "storyURL"
	keyCount       = "count"
)

var taskColumns = map[string]ui.TableColumn{
	keyId:       ui.NewTableColumn(keyId, "ID").WithAlignment(ui.TableRight),
	keyStoryId:  ui.NewTableColumn(keyStoryId, "Story ID"),
	keyProject:  ui.NewTableColumn(keyProject, "Project"),
	keyName:     ui.NewTableColumn(keyName, "Name").WithMaxWidth(40),
	keyAssignee: ui.NewTableColumn(keyAssignee, "Assignee"),
	keyReviewer: ui.NewTableColumn(keyReviewer, "Reviewer"),
	keyStatus: ui.NewTableColumn(keyStatus, "Status").WithValueFunc(
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
	keyEstimate: ui.NewTableColumn(keyEstimate, "Estimate").WithAlignment(ui.TableRight),
	keyPriority: ui.NewTableColumn(keyPriority, "Priority").WithStyleFunc(
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
	keyStoryURL:    ui.NewTableColumn(keyStoryURL, "URL"),
	keyCreatedTime: ui.NewTableColumn(keyCreatedTime, "Created"),
}

func init() {
	// Users
	TaskCmd.Flags().StringSliceP("users", "u", []string{}, "filter tasks by users (assignee or reviewer)")
	TaskCmd.Flags().StringSliceP("assignees", "a", []string{}, "filter tasks by assignees")
	TaskCmd.Flags().StringSliceP("reviewers", "r", []string{}, "filter tasks by reviewers")
	TaskCmd.MarkFlagsMutuallyExclusive("assignees", "users")
	TaskCmd.MarkFlagsMutuallyExclusive("reviewers", "users")

	// Project
	TaskCmd.Flags().StringSliceP("project", "p", []string{}, "filter by project(s)")

	// Statu VerbosityLevels
	TaskCmd.Flags().StringSliceP("status", "s", []string{}, "filter tasks by status(es) [NS, P, TBT, T, D, ND]")

	// Sprint
	TaskCmd.Flags().Var(
		flags.StringChoiceOrInt([]string{"default", "all", "backlog", "current", "next"}, "default"),
		"sprint",
		"sprint to search tasks in, by default ingnores backlog [all, default, backlog, current, <ID>]",
	)

	// Grouping
	TaskCmd.Flags().VarP(
		flags.StringChoice(
			[]string{"assignee", "project"},
			"",
		),
		"group-by",
		"g",
		"define if and how to group data [assignee]",
	)

	// Limits
	TaskCmd.Flags().Bool("all", false, "fetch all tasks")
	TaskCmd.Flags().IntP("limit", "l", 50, "limit the number of tasks to fetch")
	TaskCmd.MarkFlagsMutuallyExclusive("all", "limit")

	// Output
	keys := utils.MapKeys(taskColumns)
	defaultKeys := []string{keyStoryId, keyProject, keyName, keyAssignee, keyReviewer, keyStatus, keyEstimate, keyPriority}

	TaskCmd.Flags().Var(
		flags.StringChoiceSlice(
			keys,
			defaultKeys,
		),
		"columns",
		fmt.Sprintf("columns to show in the output table, defaults to '%s' %v", strings.Join(defaultKeys, ","), keys),
	)
	TaskCmd.Flags().Var(
		flags.StringChoiceSlice(
			keys,
			[]string{},
		),
		"add-columns",
		fmt.Sprintf("columns to add to the output table %v", keys),
	)
	TaskCmd.MarkFlagsMutuallyExclusive("columns", "add-columns")

	// Additional fields
	TaskCmd.Flags().Bool("show-url", false, "add the url of the task page to the output table")

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
		projectsList := config.Projects()
		projectsMap := config.ProjectsMap()
		timeFormat := config.DatetimeFormat()

		// Create filter
		filter := notion.TaskFilter{}

		// Assignee Flag
		if assignees, err := cmd.Flags().GetStringSlice("assignees"); err != nil {
			return err
		} else if len(assignees) != 0 {
			for _, user := range config.ParseUsers(assignees) {
				filter.Assignees = append(filter.Assignees, user.ID)
			}
		}
		// Reviewer Flag
		if reviewers, err := cmd.Flags().GetStringSlice("reviewers"); err != nil {
			return err
		} else if len(reviewers) != 0 {
			for _, user := range config.ParseUsers(reviewers) {
				filter.Reviewers = append(filter.Reviewers, user.ID)
			}
		}
		// User Flag
		if usernames, err := cmd.Flags().GetStringSlice("users"); err != nil {
			return err
		} else if len(usernames) != 0 {
			for _, user := range config.ParseUsers(usernames) {
				filter.Users = append(filter.Users, user.ID)
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
		} else if sprint == "current" || sprint == "next" {
			// Fetch sprint
			var s string
			switch sprint {
			case "current":
				s = "Current"
			case "next":
				s = "Next"
			}

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
		} else if sprintId, err := strconv.Atoi(sprint); err == nil {
			// Fetch sprint
			id := sprintId + 1
			sprintFetcher := notionClient.NewSprintFetcher(
				ctx,
				config.SprintsDatabaseID(),
				notion.SprintFilter{
					ID: &id,
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
		} else if !all {
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
		var tableStyle ui.TableStyle
		if style, err := cmd.Flags().GetString("style"); err != nil {
			return err
		} else {
			switch style {
			case "md":
				tableStyle = ui.TableStyleMarkdown
			default:
				tableStyle = ui.TableStyleDefault
			}
		}

		// Define layout
		columnKeys, err := cmd.Flags().GetStringSlice("columns")
		if err != nil {
			return err
		}

		if columnKeysToAdd, err := cmd.Flags().GetStringSlice("add-columns"); err != nil {
			return err
		} else {
			columnKeys = append(columnKeys, columnKeysToAdd...)
		}

		var columns = make([]ui.TableColumn, 0, len(columnKeys))
		for _, key := range columnKeys {
			columns = append(columns, taskColumns[key])
		}

		// Add rows
		rows := make([]ui.TableRow, 0, len(tasks))
		for _, task := range tasks {
			project := ""
			if task.ProjectID != nil {
				project = projectsMap[*task.ProjectID]
			}
			rows = append(rows, ui.TableRow{
				keyId:          task.ID,
				keyStoryId:     fmt.Sprintf("STORY-%d", task.StoryID),
				keyProject:     project,
				keyName:        task.Name,
				keyAssignee:    task.Assignee,
				keyReviewer:    task.Reviewer,
				keyStatus:      task.Status,
				keyEstimate:    fmt.Sprintf("%.1f h", task.Estimate),
				keyPriority:    task.Priority,
				keyStoryURL:    task.URL,
				keyCreatedTime: task.Created.Local().Format(timeFormat),
			})
		}

		// Render result
		table := ui.NewTable(columns).WithStyle(tableStyle).WithRows(rows)
		fmt.Println()
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

		// Grouping
		if grouping, err := cmd.Flags().GetString("group-by"); err != nil {
			return err
		} else {
			// Define grouping paramteres
			var groupKey string
			var groupTitle string
			var getGroupKeyValue func(task notion.Task) string

			switch grouping {
			case "assignee":
				groupKey = keyAssignee
				groupTitle = "Assignee"
				getGroupKeyValue = func(task notion.Task) string { return task.Assignee }
			case "project":
				groupKey = keyProject
				groupTitle = "Project"
				getGroupKeyValue = func(task notion.Task) string {
					if task.ProjectID != nil {
						return projectsMap[*task.ProjectID]
					}
					return ""
				}
			}

			// Define columns
			columns := []ui.TableColumn{
				ui.NewTableColumn(groupKey, groupTitle),
				ui.NewTableColumn(keyCount, "Count").WithAlignment(ui.TableRight),
				ui.NewTableColumn(keyEstimate, "Estimate").WithAlignment(ui.TableRight),
			}

			// Add rows
			groupingMap := make(map[string]TaskGroupingValues, 0)
			for _, task := range tasks {
				key := getGroupKeyValue(task)
				if r, ok := groupingMap[key]; ok {
					groupingMap[key] = TaskGroupingValues{
						Count: r.Count + 1,
						Hours: r.Hours + task.Estimate,
					}
				} else {
					groupingMap[key] = TaskGroupingValues{
						Count: 1,
						Hours: task.Estimate,
					}
				}
			}

			rows := make([]ui.TableRow, 0, len(groupingMap))
			for groupValue, values := range groupingMap {
				rows = append(rows, ui.TableRow{
					groupKey:    groupValue,
					keyCount:    fmt.Sprintf("%d", values.Count),
					keyEstimate: fmt.Sprintf("%.1f h", values.Hours),
				})
			}

			// Render result
			table := ui.NewTable(columns).WithStyle(tableStyle).WithRows(rows)
			fmt.Println()
			fmt.Println(table.Render())
		}

		return nil
	},
}
