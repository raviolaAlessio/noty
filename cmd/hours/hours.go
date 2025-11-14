package hours

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ravvio/easycli-ui/etable"
	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/flags"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
	"github.com/ravvio/noty/utils"
	"github.com/spf13/cobra"
)

type EntryGroupingValues struct {
	Entries int
	Hours   float64
}

var (
	keyId      = "id"
	keyDate    = "date"
	keyUser    = "user"
	keyProject = "project"
	// TODO
	// keyTask
	// keyCommission
	keyHours       = "hours"
	keyCreatedTime = "createdTime"
	keyEntries     = "entries"
)
var hoursColumns = map[string]etable.TableColumn{
	keyId:          etable.NewTableColumn(keyId, "ID"),
	keyDate:        etable.NewTableColumn(keyDate, "Date"),
	keyUser:        etable.NewTableColumn(keyUser, "User"),
	keyProject:     etable.NewTableColumn(keyProject, "Project"),
	keyHours:       etable.NewTableColumn(keyHours, "Hours").WithAlignment(etable.TableAlignmentRight),
	keyCreatedTime: etable.NewTableColumn(keyCreatedTime, "Created"),
}

func init() {
	// Users
	HoursCmd.Flags().StringSliceP("users", "u", []string{}, "filter tasks by users (assignee or reviewer)")

	// Project
	HoursCmd.Flags().StringSliceP("project", "p", []string{}, "filter by project(s)")

	// Date
	HoursCmd.Flags().VarP(
		flags.StringChoice(
			[]string{"all", "today", "yesterday"},
			"all",
		),
		"date",
		"d",
		"filter entries by date, defaults to all [all, today, yesterday]",
	)

	// Grouping
	HoursCmd.Flags().VarP(
		flags.StringChoice(
			[]string{"user", "project"},
			"",
		),
		"group-by",
		"g",
		"define if and how to group data [user]",
	)

	// Limits
	HoursCmd.Flags().Bool("all", false, "fetch all tasks")
	HoursCmd.Flags().IntP("limit", "l", 50, "limit the number of tasks to fetch")
	HoursCmd.MarkFlagsMutuallyExclusive("all", "limit")

	// Output
	keys := utils.MapKeys(hoursColumns)
	defaultKeys := []string{keyDate, keyUser, keyProject, keyHours}

	HoursCmd.Flags().Var(
		flags.StringChoiceSlice(
			keys,
			defaultKeys,
		),
		"columns",
		fmt.Sprintf("columns to show in the output table, defaults to '%s' %v", strings.Join(defaultKeys, ","), keys),
	)
	HoursCmd.Flags().Var(
		flags.StringChoiceSlice(
			keys,
			[]string{},
		),
		"add-columns",
		fmt.Sprintf("columns to add to the output table %v", keys),
	)
	HoursCmd.MarkFlagsMutuallyExclusive("columns", "add-columns")

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
		timeFormat := config.DatetimeFormat()
		dateFormat := config.DateFormat()

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
		var tableStyle etable.TableStyle
		if style, err := cmd.Flags().GetString("style"); err != nil {
			return err
		} else {
			switch style {
			case "md":
				tableStyle = etable.TableStyleMarkdown
			default:
				tableStyle = etable.TableStyleDefault
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

		var columns = make([]etable.TableColumn, 0, len(columnKeys))
		for _, key := range columnKeys {
			columns = append(columns, hoursColumns[key])
		}

		// Add rows
		rows := make([]etable.TableRow, 0, len(hoursEntries))
		for _, entry := range hoursEntries {
			project := ""
			if entry.ProjectID != nil {
				project = projectsMap[*entry.ProjectID]
			}
			rows = append(rows, etable.TableRow{
				keyId:          entry.ID,
				keyDate:        entry.Date.Format(dateFormat),
				keyProject:     project,
				keyUser:        entry.User,
				keyHours:       fmt.Sprintf("%.1f h", entry.Hours),
				keyCreatedTime: entry.Created.Local().Format(timeFormat),
			})
		}

		// Render result
		table := etable.NewTable(columns).WithStyle(tableStyle).WithRows(rows)
		fmt.Println()
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
			fd, err := os.Create(abs)
			if err != nil {
				return err
			}

			err = table.ExportCSV(fd)
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
			var getGroupKeyValue func(entry notion.HoursEntry) string

			switch grouping {
			case "user":
				groupKey = keyUser
				groupTitle = "User"
				getGroupKeyValue = func(entry notion.HoursEntry) string { return entry.User }
			case "project":
				groupKey = keyProject
				groupTitle = "Project"
				getGroupKeyValue = func(entry notion.HoursEntry) string {
					if entry.ProjectID != nil {
						return projectsMap[*entry.ProjectID]
					}
					return ""
				}
			}

			// Define columns
			columns := []etable.TableColumn{
				etable.NewTableColumn(groupKey, groupTitle),
				etable.NewTableColumn(keyEntries, "Entries").WithAlignment(etable.TableAlignmentRight),
				etable.NewTableColumn(keyHours, "Hours").WithAlignment(etable.TableAlignmentRight),
			}

			// Add rows
			groupingMap := make(map[string]EntryGroupingValues, 0)
			for _, entry := range hoursEntries {
				key := getGroupKeyValue(entry)
				if r, ok := groupingMap[key]; ok {
					groupingMap[key] = EntryGroupingValues{
						Entries: r.Entries + 1,
						Hours:   r.Hours + entry.Hours,
					}
				} else {
					groupingMap[key] = EntryGroupingValues{
						Entries: 1,
						Hours:   entry.Hours,
					}
				}
			}

			rows := make([]etable.TableRow, 0, len(groupingMap))
			for groupValue, values := range groupingMap {
				rows = append(rows, etable.TableRow{
					groupKey:   groupValue,
					keyEntries: fmt.Sprintf("%d", values.Entries),
					keyHours:   fmt.Sprintf("%.1f h", values.Hours),
				})
			}

			// Render result
			fmt.Println()
			table := etable.NewTable(columns).WithRows(rows)
			fmt.Println(table.Render())
		}

		return nil
	},
}
