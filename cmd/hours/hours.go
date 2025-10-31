package hours

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/flags"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
	"github.com/spf13/cobra"
)

// TODO UNIFY
type VerbosityLevel = int

const (
	VerbosityLevelLow     VerbosityLevel = 0
	VerbosityLevelDefault VerbosityLevel = 1
	VerbosityLevelHigh    VerbosityLevel = 2
)

func init() {
	// Users
	HoursCmd.Flags().StringSliceP("users", "u", []string{}, "filter tasks by users (assignee or reviewer)")

	// Project
	HoursCmd.Flags().StringSliceP("project", "p", []string{}, "filter by project(s)")

	// Date
	HoursCmd.Flags().VarP(
		flags.StringChoice(
			[]string{ "all", "today", "yesterday" },
			"all",
		),
		"date",
		"d",
		"filter entries by date, defaults to all [all, today, yesterday]",
	)

	// Limits
	HoursCmd.Flags().Bool("all", false, "fetch all tasks")
	HoursCmd.Flags().IntP("limit", "l", 50, "limit the number of tasks to fetch")
	HoursCmd.MarkFlagsMutuallyExclusive("all", "limit")

	// Output
	HoursCmd.Flags().VarP(
		flags.NumberChoice(
			[]int{VerbosityLevelLow, VerbosityLevelDefault, VerbosityLevelHigh},
			VerbosityLevelDefault,
		),
		"verbosity",
		"v",
		"increase or decrease amount of output fields, defaults to 1 [0, 1, 2]",
	)

	// Export
	HoursCmd.Flags().StringP("outfile", "o", "", "export result as csv")
}

var HoursCmd = &cobra.Command{
	Use:   "hours",
	Short: "fetch and analyze working hours",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		notionClient := notion.NewClient()

		// Load config
		projectsList := config.Projects()
		projectsMap := config.ProjectsMap()

		// Verbosity Flag
		verbosity, err := cmd.Flags().GetInt("verbosity")
		if err != nil {
			return err
		}
		if verbosity == 0 {
			verbosity = VerbosityLevelDefault
		}

		// Create filter
		filter := notion.HoursFilter{}

		// Users Flag
		if users, err := cmd.Flags().GetStringSlice("users"); err != nil {
			return err
		} else if len(users) != 0 {
			for _, user := range config.ParseUsers(users) {
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

		// Date Flag
		if date, err := cmd.Flags().GetString("date"); err != nil {
			return err
		} else if date == "today" {
			filter.Date = notion.HoursDateToday{}
		} else if date == "yesterday" {
			filter.Date = notion.HoursDateYesterday{}
		}

		// Create fetcher
		hoursFetcher := notionClient.NewHoursFetcher(
			ctx,
			config.HoursDatabaseID(),
			filter,
		)

		// All / Limit Flag
		if all, err := cmd.Flags().GetBool("all"); err != nil {
			return fmt.Errorf("failed request: %s", err)
		} else if !all {
			if limit, err := cmd.Flags().GetInt("limit"); err != nil {
				return err
			} else {
				hoursFetcher = hoursFetcher.WithLimit(limit)
			}
		}

		// Fetch
		hoursEntries, err := hoursFetcher.All()
		if err != nil {
			return err
		}

		// Setup table
		var (
			keyId      = "id"
			keyDate    = "date"
			keyUser    = "user"
			keyProject = "project"
			// TODO
			// keyTask
			// keyCommission
			keyHours       = "hours"
			keyCreatedTime = "created_time"
		)
		columns := []ui.TableColumn{
			ui.NewTableColumn(keyId, "ID", verbosity >= VerbosityLevelHigh),
			ui.NewTableColumn(keyDate, "Date", true),
			ui.NewTableColumn(keyUser, "User", true),
			ui.NewTableColumn(keyProject, "Project", verbosity >= VerbosityLevelDefault),
			ui.NewTableColumn(keyHours, "Hours", true).WithAlignment(ui.TableRight),
			ui.NewTableColumn(keyCreatedTime, "Created", verbosity >= VerbosityLevelHigh),
		}

		timeFormat := config.DatetimeFormat()
		dateFormat := config.DateFormat()

		// Add rows
		rows := make([]ui.TableRow, 0, len(hoursEntries))
		for _, entry := range hoursEntries {
			project := ""
			if len(entry.ProjectID) > 0 {
				project = projectsMap[entry.ProjectID[0]]
			}
			rows = append(rows, ui.TableRow{
				keyId:          entry.ID,
				keyDate:        entry.Date.Format(dateFormat),
				keyProject:     project,
				keyUser:        strings.Join(entry.User, ", "),
				keyHours:       fmt.Sprintf("%.1f h", entry.Hours),
				keyCreatedTime: entry.Created.Local().Format(timeFormat),
			})
		}
		// Render result
		table := ui.NewTable(columns).WithRows(rows)
		fmt.Println(table.Render())

		resultLog := fmt.Sprintf("\nFetched %d tasks", len(rows))
		if hoursFetcher.HasMore() {
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
