package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

type SprintFilter struct {
	ID     *int
	Status *string
}

func (self *SprintFilter) ToFilter() notionapi.Filter {
	filter := notionapi.AndCompoundFilter{}

	if self.Status != nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint status",
			Status: &notionapi.StatusFilterCondition{
				Equals: *self.Status,
			},
		})
	}

	if self.ID != nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint ID",
			UniqueId: &notionapi.UniqueIdFilterCondition{
				Equals: self.ID,
			},
		})
	}

	return filter
}

func (self *Client) NewSprintFetcher(
	ctx context.Context,
	databaseID string,
	filter SprintFilter,
) Fetcher[*SprintFetcher, Sprint] {
	client := &SprintFetcher{
		client:     self,
		databaseID: databaseID,
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

type Sprint struct {
	ID     string
	Name   string
	Status string
}

type SprintFetcher struct {
	client     *Client
	databaseID string
	filter     SprintFilter
	limit      int
	cursor     *string
}

func (self *SprintFetcher) Fetch(
	ctx context.Context,
) (FetchData[Sprint], error) {
	req := notionapi.DatabaseQueryRequest{
		PageSize: self.limit,
		Filter:   self.filter.ToFilter(),
	}
	if self.cursor != nil {
		req.StartCursor = notionapi.Cursor(*self.cursor)
	}

	res, err := self.client.client.Database.Query(
		ctx,
		notionapi.DatabaseID(self.databaseID),
		&req,
	)
	if err != nil {
		return FetchData[Sprint]{}, err
	}

	sprints := make([]Sprint, 0)
	for _, result := range res.Results {
		sprints = append(sprints, Sprint{
			ID:     string(result.ID),
			Name:   ParseTitle(result.Properties["Sprint name"]),
			Status: ParseStatus(result.Properties["Sprint status"]),
		})
	}

	fd := FetchData[Sprint]{
		NextToken: nil,
		Data:      sprints,
	}
	if res.HasMore {
		cursor := res.NextCursor.String()
		fd.NextToken = &cursor
	}
	return fd, nil
}

func (self *SprintFetcher) RequestLimit() int {
	return self.limit
}

func (self *SprintFetcher) SetRequestLimit(limit int) {
	self.limit = limit
}

func (self *SprintFetcher) SetNextToken(cursor *string) {
	self.cursor = cursor
}
