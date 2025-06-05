package notion

import (
	"fmt"

	"github.com/jomei/notionapi"
)

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

func ParseStatus(p notionapi.Property) string {
	return  p.(*notionapi.StatusProperty).Status.Name
}

func ParsePeople(p notionapi.Property) []string {
	users := p.(*notionapi.PeopleProperty).People
	result := make([]string, 0, len(users))
	for _, user := range users {
		result = append(result, user.Name)
	}
	return result
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

func ParseDate(p notionapi.Property) string {
	start := p.(*notionapi.DateProperty).Date.Start
	end := p.(*notionapi.DateProperty).Date.End
	if start != nil && end != nil {
		return fmt.Sprintf("%s - %s", start.String(), end.String())
	} else if start != nil {
		return start.String()
	} else {
		return end.String()
	}
}
