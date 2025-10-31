package notion

import (
	"context"
	"time"

	"github.com/jomei/notionapi"
)

type HoursDateFilter interface {
	ToFilter() notionapi.Filter
}

type HoursDateToday struct{}

func (dateFilter HoursDateToday) ToFilter() notionapi.Filter {
	now := notionapi.Date(time.Now().Round(24 * time.Hour))

	return notionapi.PropertyFilter{
		Property: "data",
		Date: &notionapi.DateFilterCondition{
			Equals: &now,
		},
	}
}

type HoursDateYesterday struct{}

func (dateFilter HoursDateYesterday) ToFilter() notionapi.Filter {
	now := notionapi.Date(time.Now().Add(-24 * time.Hour).Round(24 * time.Hour))

	return notionapi.PropertyFilter{
		Property: "data",
		Date: &notionapi.DateFilterCondition{
			Equals: &now,
		},
	}
}

type HoursFilter struct {
	Projects []string
	Users    []string
	Date     HoursDateFilter
}

func (hoursFilter *HoursFilter) ToFilter() notionapi.Filter {
	filter := notionapi.AndCompoundFilter{}

	if len(hoursFilter.Projects) > 0 {
		projectsFilter := notionapi.OrCompoundFilter{}
		for _, project := range hoursFilter.Projects {
			projectsFilter = append(projectsFilter, notionapi.PropertyFilter{
				Property: "progetto",
				Relation: &notionapi.RelationFilterCondition{
					Contains: project,
				},
			})
		}
		filter = append(filter, projectsFilter)
	}

	if len(hoursFilter.Users) > 0 {
		userFilter := notionapi.OrCompoundFilter{}
		for _, u := range hoursFilter.Users {
			userFilter = append(
				userFilter,
				notionapi.PropertyFilter{
					Property: "codeployer",
					People: &notionapi.PeopleFilterCondition{
						Contains: u,
					},
				},
			)
		}
		filter = append(filter, userFilter)
	}

	if hoursFilter.Date != nil {
		filter = append(filter, hoursFilter.Date.ToFilter())
	}

	return filter
}

type HoursEntry struct {
	ID           string
	Created      time.Time
	User         []string
	ProjectID    []string
	TaskID       *string
	CommissionID []string
	Date         time.Time
	Hours        float64
}

func parseHoursEntryPage(p notionapi.Page) (HoursEntry, error) {
	var taskId *string
	if teaskRel := ParseRelation(p.Properties["task"]); len(teaskRel) > 0 {
		taskId = &teaskRel[0]
	}

	return HoursEntry{
		ID:           p.ID.String(),
		Created:      p.CreatedTime,
		User:         ParsePeople(p.Properties["codeployer"]),
		ProjectID:    ParseRelation(p.Properties["progetto"]),
		TaskID:       taskId,
		CommissionID: ParseRelation(p.Properties["commessa"]),
		Date:         ParseDate(p.Properties["data"]).Start,
		Hours:        ParseNumber(p.Properties["ore"]),
	}, nil
}

func (client *Client) NewHoursFetcher(
	ctx context.Context,
	databaseId string,
	filter HoursFilter,
) Fetcher[*HoursFetcher, HoursEntry] {
	fetcher := &HoursFetcher{
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

type HoursFetcher struct {
	client     *Client
	databaseId string
	filter     HoursFilter
	limit      int
	cursor     *string
}

func (fetcher *HoursFetcher) Fetch(
	ctx context.Context,
) (FetchData[HoursEntry], error) {
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
		return FetchData[HoursEntry]{}, err
	}

	hoursEntries := make([]HoursEntry, 0, len(res.Results))
	for _, result := range res.Results {
		hoursEntry, err := parseHoursEntryPage(result)
		if err != nil {
			return FetchData[HoursEntry]{}, err
		}
		hoursEntries = append(hoursEntries, hoursEntry)
	}

	fd := FetchData[HoursEntry]{
		NextToken: nil,
		Data:      hoursEntries,
	}
	if res.HasMore {
		cursor := res.NextCursor.String()
		fd.NextToken = &cursor
	}
	return fd, nil
}

func (fetcher *HoursFetcher) RequestLimit() int {
	return fetcher.limit
}

func (fetcher *HoursFetcher) SetRequestLimit(limit int) {
	fetcher.limit = limit
}

func (fetcher *HoursFetcher) SetNextToken(cursor *string) {
	fetcher.cursor = cursor
}
