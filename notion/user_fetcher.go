package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

func (client *Client) NewUserFetcher(
	ctx context.Context,
	excludeBots bool,
) Fetcher[*UserFetcher, NotionUser] {
	fetcher := &UserFetcher{
		client:      client,
		limit:       100,
		cursor:      nil,
		excludeBots: excludeBots,
	}
	return NewFetcher(
		ctx,
		fetcher,
		100,
	)
}

type NotionUser struct {
	ID   string
	Name string
}

type UserFetcher struct {
	client      *Client
	limit       int
	cursor      *string
	excludeBots bool
}

func (fetcher *UserFetcher) Fetch(
	ctx context.Context,
) (FetchData[NotionUser], error) {
	req := notionapi.Pagination{
		PageSize: fetcher.limit,
	}
	if fetcher.cursor != nil {
		req.StartCursor = notionapi.Cursor(*fetcher.cursor)
	}

	res, err := fetcher.client.client.User.List(
		ctx,
		&req,
	)
	if err != nil {
		return FetchData[NotionUser]{}, err
	}

	users := make([]NotionUser, 0)
	for _, result := range res.Results {
		if fetcher.excludeBots && result.Bot != nil {
			continue
		}
		users = append(users, NotionUser{
			ID:   string(result.ID),
			Name: result.Name,
		})
	}

	fd := FetchData[NotionUser]{
		NextToken: nil,
		Data:      users,
	}
	if res.HasMore {
		cursor := res.NextCursor.String()
		fd.NextToken = &cursor
	}
	return fd, nil
}

func (fetcher *UserFetcher) RequestLimit() int {
	return fetcher.limit
}

func (fetcher *UserFetcher) SetRequestLimit(limit int) {
	fetcher.limit = limit
}

func (fetcher *UserFetcher) SetNextToken(cursor *string) {
	fetcher.cursor = cursor
}
