package notion

import (
	"context"
	"time"

	"github.com/jomei/notionapi"
)

const (
	StatusNotStarted = "Not Started"
	StatusInProgress = "In Progress"
	StatusToBeTested = "To Be Tested"
	StatusInTesting  = "In Testing"
	StatusDone       = "Done"
	StatusNotDone    = "Not Done"
)

type TaskSprintFilter interface {
	ToFilter() notionapi.Filter
}

type TaskSprintNoBacklog struct{}

func (sprintFilter TaskSprintNoBacklog) ToFilter() notionapi.Filter {
	return notionapi.PropertyFilter{
		Property: "Sprint",
		Relation: &notionapi.RelationFilterCondition{
			IsNotEmpty: true,
		},
	}
}

type TaskSprintOnlyBacklog struct{}

func (sprintFilter TaskSprintOnlyBacklog) ToFilter() notionapi.Filter {
	return notionapi.PropertyFilter{
		Property: "Sprint",
		Relation: &notionapi.RelationFilterCondition{
			IsEmpty: true,
		},
	}
}

type TaskSprintByID struct {
	ID string
}

func (sprintFilter TaskSprintByID) ToFilter() notionapi.Filter {
	return notionapi.PropertyFilter{
		Property: "Sprint",
		Relation: &notionapi.RelationFilterCondition{
			Contains: sprintFilter.ID,
		},
	}
}

type TaskSprintByIDs struct {
	SprintIDs []string
}

func (sprintFilter TaskSprintByIDs) ToFilter() notionapi.Filter {
	filter := notionapi.OrCompoundFilter{}
	for _, id := range sprintFilter.SprintIDs {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint",
			Relation: &notionapi.RelationFilterCondition{
				Contains: id,
			},
		})
	}
	return filter
}

type TaskFilter struct {
	Projects  []string
	Users     []string
	Assignees []string
	Reviewers []string
	Statuses  []string
	Sprint    TaskSprintFilter
	Estimate  string
}

func (taskFilter *TaskFilter) ToFilter() notionapi.Filter {
	filter := notionapi.AndCompoundFilter{}

	if len(taskFilter.Projects) > 0 {
		projectsFilter := notionapi.OrCompoundFilter{}
		for _, project := range taskFilter.Projects {
			projectsFilter = append(projectsFilter, notionapi.PropertyFilter{
				Property: "Project",
				Relation: &notionapi.RelationFilterCondition{
					Contains: project,
				},
			})
		}
		filter = append(filter, projectsFilter)
	}

	if len(taskFilter.Users) > 0 {
		userFilter := notionapi.OrCompoundFilter{}
		for _, u := range taskFilter.Users {
			userFilter = append(
				userFilter,
				notionapi.PropertyFilter{
					Property: "Assignee",
					People: &notionapi.PeopleFilterCondition{
						Contains: u,
					},
				},
				notionapi.PropertyFilter{
					Property: "Reviewer",
					People: &notionapi.PeopleFilterCondition{
						Contains: u,
					},
				},
			)
		}
		filter = append(filter, userFilter)
	}

	if len(taskFilter.Assignees) > 0 && len(taskFilter.Users) == 0 {
		userFilter := notionapi.OrCompoundFilter{}
		for _, u := range taskFilter.Assignees {
			userFilter = append(
				userFilter,
				notionapi.PropertyFilter{
					Property: "Assignee",
					People: &notionapi.PeopleFilterCondition{
						Contains: u,
					},
				},
			)
		}
		filter = append(filter, userFilter)
	}

	if len(taskFilter.Reviewers) > 0 && len(taskFilter.Users) == 0 {
		userFilter := notionapi.OrCompoundFilter{}
		for _, u := range taskFilter.Reviewers {
			userFilter = append(
				userFilter,
				notionapi.PropertyFilter{
					Property: "Reviewer",
					People: &notionapi.PeopleFilterCondition{
						Contains: u,
					},
				},
			)
		}
		filter = append(filter, userFilter)
	}

	if len(taskFilter.Statuses) > 0 {
		statusFilter := notionapi.OrCompoundFilter{}
		for _, status := range taskFilter.Statuses {
			statusFilter = append(statusFilter, notionapi.PropertyFilter{
				Property: "Status",
				Status: &notionapi.StatusFilterCondition{
					Equals: status,
				},
			})
		}
		filter = append(filter, statusFilter)
	}

	if taskFilter.Sprint != nil {
		filter = append(filter, (taskFilter.Sprint).ToFilter())
	}

	return filter
}

type Task struct {
	ID        string
	StoryID   int
	Name      string
	Assignee  string
	Reviewer  string
	Status    string
	Priority  string
	ProjectID *string
	Created   time.Time
	Estimate  float64
	SprintID  string
	URL       string
}

func parseTaskPage(p notionapi.Page) (Task, error) {
	return Task{
		ID:        p.ID.String(),
		StoryID:   ParseUniqueID(p.Properties["Story ID"]),
		Name:      ParseTitle(p.Properties["Task name"]),
		Status:    ParseStatus(p.Properties["Status"]),
		Assignee:  ParseUserName(p.Properties["Assignee"], "-"),
		Reviewer:  ParseUserName(p.Properties["Reviewer"], "-"),
		Priority:  ParseSelect(p.Properties["Priority"]),
		ProjectID: OneOrNil(ParseRelation(p.Properties["Project"])),
		Created:   p.CreatedTime,
		Estimate:  ParseNumber(p.Properties["estimate hours"]),
		SprintID:  ParseRelation(p.Properties["Sprint"])[0],
		URL:       p.URL,
	}, nil
}

func (client *Client) NewTaskFetcher(
	ctx context.Context,
	databaseId string,
	filter TaskFilter,
) Fetcher[*TaskFetcher, Task] {
	fetcher := &TaskFetcher{
		client:     client,
		databaseId: databaseId,
		filter:     filter,
		limit:      100,
		cursor:     nil,
	}
	return NewFetcher(
		ctx,
		fetcher,
		100,
	)
}

type TaskFetcher struct {
	client     *Client
	databaseId string
	filter     TaskFilter
	limit      int
	cursor     *string
}

func (fetcher *TaskFetcher) Fetch(
	ctx context.Context,
) (FetchData[Task], error) {
	req := &notionapi.DatabaseQueryRequest{
		Filter:   fetcher.filter.ToFilter(),
		PageSize: int(fetcher.limit),
	}
	if fetcher.cursor != nil {
		req.StartCursor = notionapi.Cursor(*fetcher.cursor)
	}

	res, err := fetcher.client.client.Database.Query(
		ctx,
		notionapi.DatabaseID(fetcher.databaseId),
		req,
	)
	if err != nil {
		return FetchData[Task]{}, err
	}

	tasks := make([]Task, 0, len(res.Results))
	for _, result := range res.Results {
		task, err := parseTaskPage(result)
		if err != nil {
			return FetchData[Task]{}, err
		}
		tasks = append(tasks, task)
	}

	fd := FetchData[Task]{
		NextToken: nil,
		Data:      tasks,
	}
	if res.HasMore {
		cursor := res.NextCursor.String()
		fd.NextToken = &cursor
	}
	return fd, nil
}

func (fetcher *TaskFetcher) RequestLimit() int {
	return fetcher.limit
}

func (fetcher *TaskFetcher) SetRequestLimit(limit int) {
	fetcher.limit = limit
}

func (fetcher *TaskFetcher) SetNextToken(cursor *string) {
	fetcher.cursor = cursor
}
