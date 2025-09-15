package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

func (client *Client) NewProjectFetcher(
	ctx context.Context,
	databaseID string,
) Fetcher[*ProjectFetcher, Project] {
	fetcher := &ProjectFetcher{
		client:     client,
		limit:      100,
		cursor:     nil,
		databaseID: databaseID,
	}
	return NewFetcher(
		ctx,
		fetcher,
		100,
	)
}

type Project struct {
	ID   string
	Name string
}

type ProjectFetcher struct {
	client     *Client
	limit      int
	cursor     *string
	databaseID string
}

func (fetcher *ProjectFetcher) Fetch(
	ctx context.Context,
) (FetchData[Project], error) {
	req := notionapi.DatabaseQueryRequest{
		PageSize: fetcher.limit,
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
		return FetchData[Project]{}, err
	}

	projects := make([]Project, 0)
	for _, result := range res.Results {
		projects = append(projects, Project{
			ID:   string(result.ID),
			Name: ParseTitle(result.Properties["Project name"]),
		})
	}

	fd := FetchData[Project]{
		NextToken: nil,
		Data:      projects,
	}
	if res.HasMore {
		cursor := res.NextCursor.String()
		fd.NextToken = &cursor
	}
	return fd, nil
}

func (fetcher *ProjectFetcher) RequestLimit() int {
	return fetcher.limit
}

func (fetcher *ProjectFetcher) SetRequestLimit(limit int) {
	fetcher.limit = limit
}

func (fetcher *ProjectFetcher) SetNextToken(cursor *string) {
	fetcher.cursor = cursor
}
