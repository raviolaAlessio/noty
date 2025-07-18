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
)

type SprintType int

const (
	SprintTypeAll = iota
	SprintTypeNoBacklog
	SprintTypeOnlyBacklog
)

type TaskFilter struct {
	Projects []string
	User     *string
	Assignee *string
	Reviewer *string
	Statuses []string
	Sprint   SprintType
}

func (self *TaskFilter) ToFilter() notionapi.Filter {
	filter := notionapi.AndCompoundFilter{}

	if len(self.Projects) > 0 {
		projectsFilter := notionapi.OrCompoundFilter{}
		for _, project := range self.Projects {
			projectsFilter = append(projectsFilter, notionapi.PropertyFilter{
				Property: "Project",
				Relation: &notionapi.RelationFilterCondition{
					Contains: project,
				},
			})
		}
		filter = append(filter, projectsFilter)
	}

	if self.User != nil {
		userFilter := notionapi.OrCompoundFilter{
			notionapi.PropertyFilter{
				Property: "Assignee",
				People: &notionapi.PeopleFilterCondition{
					Contains: *self.User,
				},
			},
			notionapi.PropertyFilter{
				Property: "Reviewer",
				People: &notionapi.PeopleFilterCondition{
					Contains: *self.User,
				},
			},
		}
		filter = append(filter, userFilter)
	}

	if self.Assignee != nil && self.User == nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Assignee",
			People: &notionapi.PeopleFilterCondition{
				Contains: *self.Assignee,
			},
		})
	}

	if self.Reviewer != nil && self.User == nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Reviewer",
			People: &notionapi.PeopleFilterCondition{
				Contains: *self.Reviewer,
			},
		})
	}

	if len(self.Statuses) > 0 {
		statusFilter := notionapi.OrCompoundFilter{}
		for _, status := range self.Statuses {
			statusFilter = append(statusFilter, notionapi.PropertyFilter{
				Property: "Status",
				Status: &notionapi.StatusFilterCondition{
					Equals: status,
				},
			})
		}
		filter = append(filter, statusFilter)
	}

	switch self.Sprint {
	case SprintTypeNoBacklog:
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint",
			Relation: &notionapi.RelationFilterCondition{
				IsNotEmpty: true,
			},
		})
		break
	case SprintTypeOnlyBacklog:
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint",
			Relation: &notionapi.RelationFilterCondition{
				IsEmpty: true,
			},
		})
		break
	default:
		break
	}

	return filter
}

type Task struct {
	ID        string
	Name      string
	Assignee  []string
	Reviewer  []string
	Status    string
	Priority  string
	ProjectID []string
	Created   time.Time
	// TODO: missing authorization
	// Backlog   bool
}

func parseTaskPage(p notionapi.Page) Task {
	return Task{
		ID:        p.ID.String(),
		Name:      ParseTitle(p.Properties["Task name"]),
		Status:    ParseStatus(p.Properties["Status"]),
		Assignee:  ParsePeople(p.Properties["Assignee"]),
		Reviewer:  ParsePeople(p.Properties["Reviewer"]),
		Priority:  ParseSelect(p.Properties["Priority"]),
		ProjectID: ParseRelation(p.Properties["Project"]),
		Created:   p.CreatedTime,
	}
}

func (self *Client) NewTaskFetcher(
	ctx context.Context,
	databaseId string,
	filter TaskFilter,
) Fetcher[*TaskFetcher, Task] {
	client := &TaskFetcher{
		client:     self,
		databaseId: databaseId,
		filter:     filter,
		limit:      100,
		cursor:     nil,
	}
	return NewFetcher(
		ctx,
		client,
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

func (self *TaskFetcher) Fetch(
	ctx context.Context,
) (FetchData[Task], error) {
	req := &notionapi.DatabaseQueryRequest{
		Filter:   self.filter.ToFilter(),
		PageSize: int(self.limit),
	}
	if self.cursor != nil {
		req.StartCursor = notionapi.Cursor(*self.cursor)
	}

	res, err := self.client.client.Database.Query(
		ctx,
		notionapi.DatabaseID(self.databaseId),
		req,
	)
	if err != nil {
		return FetchData[Task]{}, err
	}

	tasks := make([]Task, 0, len(res.Results))
	for _, result := range res.Results {
		tasks = append(tasks, parseTaskPage(result))
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

func (self *TaskFetcher) RequestLimit() int {
	return self.limit
}

func (self *TaskFetcher) SetRequestLimit(limit int) {
	self.limit = limit
}

func (self *TaskFetcher) SetNextToken(cursor *string) {
	self.cursor = cursor
}
