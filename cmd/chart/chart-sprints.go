package chart

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/ravvio/noty/config"
	"github.com/ravvio/noty/flags"
	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
	"github.com/spf13/cobra"
)

func init() {
	ChartSprints.Flags().VarP(
		flags.StringChoice([]string{"hours"}),
		"y-axis",
		"y",
		"values to show on the y-axis [hours]",
	)

	ChartSprints.Flags().StringSliceP(
		"users",
		"u",
		[]string{},
		"users to add to the chart",
	)
	ChartSprints.MarkFlagsOneRequired(
		"users",
	)

	// Output
	ChartSprints.Flags().StringP("outfile", "o", "chart.html", "name of the generated html file")
}

type chartAxis = int

const (
	chartAxisHours chartAxis = iota
)

var ChartSprints = &cobra.Command{
	Use:   "sprints",
	Short: "create a bar chart with sprints on the X axis",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		notionClient := notion.NewClient()

		// Load config
		usersList := config.Users()

		// Selected Y axis value
		yAxisFlag, err := cmd.Flags().GetString("y-axis")
		if err != nil {
			return err
		}
		var chartYAxis chartAxis
		switch yAxisFlag {
		case "hours":
			chartYAxis = chartAxisHours
		}

		// Selected users
		usersFlag, err := cmd.Flags().GetStringSlice("users")
		if err != nil {
			return err
		}
		var users []notion.NotionUser = make([]notion.NotionUser, 0)
		for _, uf := range usersFlag {
			var found *notion.NotionUser;
			for _, u := range usersList {
				if strings.Contains(strings.ToLower(u.Name), strings.ToLower(uf)) {
					found = &u
					break
				}
			}
			if found == nil {
				ui.PrintlnfWarn("no user found for '%s'", uf)
			} else {
				users = append(users, *found)
			}
		}

		startSprint := 65
		endSprint := 70

		if startSprint >= endSprint {
			return fmt.Errorf("starting sprint '%d' must be < than end sprint '%d'", startSprint, endSprint)
		}

		// Load sprints
		var sprints []notion.Sprint = nil
		_, err = ui.Spin(
			"Loading sprints",
			func() error {
				// TODO only load required sprints
				sprintFetcher := notionClient.NewSprintFetcher(
					ctx,
					config.SprintsDatabaseID(),
					notion.SprintFilter{},
				)
				sprints, err = sprintFetcher.All()
				return nil
			},
		)
		if err != nil {
			return err
		}

		// Find tasks and aggregate
		var results [][]opts.BarData = make([][]opts.BarData, len(users))

		for userIndex, user := range users {
			_, err = ui.Spin(
				fmt.Sprintf("Loading and aggregating tasks for user %s", user.Name),
				func() error {
					for sprintIndex := startSprint; sprintIndex <= endSprint; sprintIndex++ {
						// Select sprint
						var sprint *notion.Sprint
						for _, sp := range sprints {
							if sp.Name == fmt.Sprintf("Sprint %d", sprintIndex+1) {
								sprint = &sp
							}
						}
						if sprint == nil {
							return fmt.Errorf("could not find sprint %d", sprintIndex)
						}

						// Find tasks
						taskFetcher := notionClient.NewTaskFetcher(
							ctx,
							config.TasksDatabaseID(),
							notion.TaskFilter{
								User: &user.ID,
								Sprint: notion.TaskSprintByID{
									ID: sprint.ID,
								},
							},
						)
						tasks, err := taskFetcher.All()
						if err != nil {
							return err
						}

						value := 0.0

						switch chartYAxis {
						case chartAxisHours:
							for _, task := range tasks {
								value += task.Estimate
							}
						}

						results[userIndex] = append(results[userIndex], opts.BarData{
							Value: value,
						})
					}
					return nil
				},
			)
			if err != nil {
				return err
			}
		}

		// Create and render chart
		chart := charts.NewBar()
		chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
			Title: "Sprint hours",
		}))

		xAxis := make([]string, 0)
		for i := startSprint; i <= endSprint; i += 1 {
			xAxis = append(xAxis, fmt.Sprintf("Sprint %d", i))
		}

		for userIndex, user := range users {
			chart.
				SetXAxis(xAxis).
				AddSeries(user.Name, results[userIndex])
		}

		if outfile, err := cmd.Flags().GetString("outfile"); err != nil {
			return err
		} else if outfile != "" {
			abs, err := filepath.Abs(outfile)
			if err != nil {
				return err
			}
			f, err := os.Create(abs)
			if err != nil {
				return err
			}
			err = chart.Render(f)
			if err != nil {
				return err
			}
		}

		return nil
	},
}
