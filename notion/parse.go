package notion

import (
	"fmt"
	"time"

	"github.com/jomei/notionapi"
)

func ParseUniqueID(p notionapi.Property) int {
	title := p.(*notionapi.UniqueIDProperty).UniqueID
	return title.Number
}

func ParseTitle(p notionapi.Property) string {
	title := p.(*notionapi.TitleProperty).Title
	result := ""
	for _, text := range title {
		result += text.PlainText
	}
	return result
}

func ParseRichText(p notionapi.Property) string {
	richtext := p.(*notionapi.RichTextProperty).RichText
	result := ""
	for _, text := range richtext {
		result += text.PlainText
	}
	return result
}

func ParseNumber(p notionapi.Property) float64 {
	return p.(*notionapi.NumberProperty).Number
}

func ParseStatus(p notionapi.Property) string {
	return p.(*notionapi.StatusProperty).Status.Name
}

func ParsePeople(p notionapi.Property) []string {
	users := p.(*notionapi.PeopleProperty).People
	result := make([]string, 0, len(users))
	for _, user := range users {
		result = append(result, user.Name)
	}
	return result
}

func ParseUserName(p notionapi.Property, fallback string) string {
	users := p.(*notionapi.PeopleProperty).People
	if len(users) == 0 {
		return fallback
	}
	return users[0].Name
}

func ParseSelect(p notionapi.Property) string {
	return p.(*notionapi.SelectProperty).Select.Name
}

func ParseRelation(p notionapi.Property) []string {
	rel := p.(*notionapi.RelationProperty).Relation
	res := make([]string, 0, len(rel))
	for _, r := range rel {
		res = append(res, r.ID.String())
	}
	return res
}

type DateRange struct {
	Start time.Time
	End   time.Time
}

func ParseDate(p notionapi.Property) DateRange {
	startDate, err := time.Parse(
		time.RFC3339,
		p.(*notionapi.DateProperty).Date.Start.String(),
	)
	if err != nil {
		startDate = time.Now()
	}
	endDate, err := time.Parse(
		time.RFC3339,
		p.(*notionapi.DateProperty).Date.Start.String(),
	)
	if err != nil {
		startDate = time.Now()
	}

	return DateRange{
		Start: startDate,
		End:   endDate,
	}
}

func ParseRollup(p notionapi.Property) string {
	value := p.(*notionapi.RollupProperty).Rollup.Number
	return fmt.Sprintf("Rollup %f", value)
}
