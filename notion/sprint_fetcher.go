package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

type SprintFilter struct {
	ID     *int
	Status *string
}

func (sprintFilter *SprintFilter) ToFilter() notionapi.Filter {
	filter := notionapi.AndCompoundFilter{}

	if sprintFilter.Status != nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint status",
			Status: &notionapi.StatusFilterCondition{
				Equals: *sprintFilter.Status,
			},
		})
	}

	if sprintFilter.ID != nil {
		filter = append(filter, notionapi.PropertyFilter{
			Property: "Sprint ID",
			UniqueId: &notionapi.UniqueIdFilterCondition{
				Equals: sprintFilter.ID,
			},
		})
	}

	return filter
}

func (client *Client) NewSprintFetcher(
	ctx context.Context,
	databaseID string,
	filter SprintFilter,
) Fetcher[*SprintFetcher, Sprint] {
	fetcher := &SprintFetcher{
		client:     client,
		databaseID: databaseID,
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

type Sprint struct {
	ID       string
	SprintID int
	Name     string
	Status   string
}

type SprintFetcher struct {
	client     *Client
	databaseID string
	filter     SprintFilter
	limit      int
	cursor     *string
}

func (fetcher *SprintFetcher) Fetch(
	ctx context.Context,
) (FetchData[Sprint], error) {
	req := notionapi.DatabaseQueryRequest{
		PageSize: fetcher.limit,
		Filter:   fetcher.filter.ToFilter(),
	}
	if fetcher.cursor != nil {
		req.StartCursor = notionapi.Cursor(*fetcher.cursor)
	}

	res, err := fetcher.client.client.Database.Query(
		ctx,
		notionapi.DatabaseID(fetcher.databaseID),
		&req,
	)
	if err != nil {
		return FetchData[Sprint]{}, err
	}

	sprints := make([]Sprint, 0)
	for _, result := range res.Results {
		sprints = append(sprints, Sprint{
			ID:       result.ID.String(),
			SprintID: ParseUniqueID(result.Properties["Sprint ID"]),
			Name:     ParseTitle(result.Properties["Sprint name"]),
			Status:   ParseStatus(result.Properties["Sprint status"]),
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

func (fetcher *SprintFetcher) RequestLimit() int {
	return fetcher.limit
}

func (fetcher *SprintFetcher) SetRequestLimit(limit int) {
	fetcher.limit = limit
}

func (fetcher *SprintFetcher) SetNextToken(cursor *string) {
	fetcher.cursor = cursor
}
