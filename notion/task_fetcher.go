package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

const (
	StatusNotStarted = "Not Started"
	StatusInProgress = "In Progress"
	StatusToBeTested = "To Be Tested"
	StatusInTesting  = "In Testing"
	StatusDone       = "Done"
)

type TaskFilter struct {
	Project  *string
	Assignee *string
	Statuses []string
}

func (self *TaskFilter) ToFilter() notionapi.Filter {
	filter := notionapi.AndCompoundFilter{}

	if self.Project != nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Project",
			RichText: &notionapi.TextFilterCondition{
				Equals: *self.Project,
			},
		})
	}

	if self.Assignee != nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Assignee",
			People: &notionapi.PeopleFilterCondition{
				Contains: *self.Assignee,
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
