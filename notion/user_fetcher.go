package notion

import (
	"context"

	"github.com/jomei/notionapi"
)

func (self *Client) NewUserFetcher(
	ctx context.Context,
	excludeBots bool,
) Fetcher[*UserFetcher, NotionUser] {
	client := &UserFetcher{
		client: self,
		limit:  100,
		cursor: nil,
		excludeBots: excludeBots,
	}
	return NewFetcher(
		ctx,
		client,
		100,
	)
}

type NotionUser struct {
	ID string
	Name string
}

type UserFetcher struct {
	client *Client
	limit  int
	cursor *string
	excludeBots bool
}

func (self *UserFetcher) Fetch(
	ctx context.Context,
) (FetchData[NotionUser], error) {
	req := notionapi.Pagination{
		PageSize:    self.limit,
	}
	if self.cursor != nil {
		req.StartCursor = notionapi.Cursor(*self.cursor)
	}

	res, err := self.client.client.User.List(
		ctx,
		&req,
	)
	if err != nil {
		return FetchData[NotionUser]{}, err
	}

	users := make([]NotionUser, 0)
	for _, result := range res.Results {
		if (self.excludeBots && result.Bot != nil) {
			continue
		}
		users = append(users, NotionUser{
			ID: string(result.ID),
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

func (self *UserFetcher) RequestLimit() int {
	return self.limit
}

func (self *UserFetcher) SetRequestLimit(limit int) {
	self.limit = limit
}

func (self *UserFetcher) SetNextToken(cursor *string) {
	self.cursor = cursor
}
