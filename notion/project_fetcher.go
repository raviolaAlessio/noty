package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

func (self *Client) NewProjectFetcher(
	ctx context.Context,
	databaseID string,
) Fetcher[*ProjectFetcher, Project] {
	client := &ProjectFetcher{
		client:     self,
		limit:      100,
		cursor:     nil,
		databaseID: databaseID,
	}
	return NewFetcher(
		ctx,
		client,
		100,
	)
}

type Project struct {
	ID    string
	Name string
}

type ProjectFetcher struct {
	client     *Client
	limit      int
	cursor     *string
	databaseID string
}

func (self *ProjectFetcher) Fetch(
	ctx context.Context,
) (FetchData[Project], error) {
	req := notionapi.DatabaseQueryRequest{
		PageSize: self.limit,
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
		return FetchData[Project]{}, err
	}

	projects := make([]Project, 0)
	for _, result := range res.Results {
		projects = append(projects, Project{
			ID:    string(result.ID),
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

func (self *ProjectFetcher) RequestLimit() int {
	return self.limit
}

func (self *ProjectFetcher) SetRequestLimit(limit int) {
	self.limit = limit
}

func (self *ProjectFetcher) SetNextToken(cursor *string) {
	self.cursor = cursor
}
