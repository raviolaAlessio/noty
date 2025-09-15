package chart

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
		"assignees",
		"a",
		[]string{},
		"assignees to add to the chart",
	)
	ChartSprints.MarkFlagsOneRequired(
		"assignees",
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

		// Selected Y axis value
		yAxisFlag, err := cmd.Flags().GetString("y-axis")
		if err != nil {
			return err
		}
		var chartYAxis chartAxis
		switch yAxisFlag {
		case "hours":
			chartYAxis = chartAxisHours
		default:
			chartYAxis = chartAxisHours
		}

		// Selected assignees
		assigneesFlag, err := cmd.Flags().GetStringSlice("assignees")
		if err != nil {
			return err
		}
		assignees := config.ParseUsers(assigneesFlag)
		assigneesIDs := make([]string, len(assignees))
		for i, u := range assignees {
			assigneesIDs[i] = u.ID
		}

		startSprint := 65
		endSprint := 70

		if startSprint >= endSprint {
			return fmt.Errorf("starting sprint '%d' must be < than end sprint '%d'", startSprint, endSprint)
		}

		// Load sprints
		sprints := make([]notion.Sprint, 0)
		_, err = ui.Spin(
			"Loading sprints",
			func() error {
				// TODO only load required sprints
				sprintFetcher := notionClient.NewSprintFetcher(
					ctx,
					config.SprintsDatabaseID(),
					notion.SprintFilter{},
				)
				allSprints, err := sprintFetcher.All()
				if err != nil {
					return err
				}
				for _, s := range allSprints {
					if startSprint <= s.SprintID && s.SprintID <= endSprint {
						sprints = append(sprints, s)
					}
				}
				return nil
			},
		)
		if err != nil {
			return err
		}
		sprintMap := make(map[string]int)
		sprintIDs := make([]string, len(sprints))
		for i, s := range sprints {
			sprintIDs[i] = s.ID
			sprintMap[s.ID] = i
		}

		// Find tasks and aggregate
		results := make(map[string][]float64)
		for _, u := range assignees {
			results[u.Name] = make([]float64, len(sprints))
			for i := range len(sprints) {
				results[u.Name][i] = 0.0
			}
		}

		_, err = ui.Spin(
			"Loading and aggregating tasks",
			func() error {
				// Find tasks
				taskFetcher := notionClient.NewTaskFetcher(
					ctx,
					config.TasksDatabaseID(),
					notion.TaskFilter{
						Assignees: assigneesIDs,
						Sprint: notion.TaskSprintByIDs{
							SprintIDs: sprintIDs,
						},
					},
				)
				tasks, err := taskFetcher.All()
				if err != nil {
					return err
				}

				// Aggregate
				switch chartYAxis {
				case chartAxisHours:
					for _, task := range tasks {
						results[task.Assignee[0]][sprintMap[task.SprintID]] += task.Estimate
					}
				}

				return nil
			},
		)
		if err != nil {
			return err
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
		chart.SetXAxis(xAxis)

		for username, datapoints := range results {
			serie := make([]opts.BarData, len(datapoints))
			for i, d := range datapoints {
				serie[i] = opts.BarData{
					Value: d,
				}
			}
			chart.AddSeries(username, serie)
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
