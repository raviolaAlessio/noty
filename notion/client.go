package notion

import (
	"os"

	"github.com/jomei/notionapi"
)

var (
	token = os.Getenv("NOTION_API_KEY")
)

type Client struct {
	client *notionapi.Client
}

func NewClient() *Client {
	client := notionapi.NewClient(notionapi.Token(token))
	return &Client{
		client: client,
	}
}

