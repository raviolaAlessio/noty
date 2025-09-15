package config

import (
	"strings"

	"github.com/ravvio/noty/notion"
	"github.com/ravvio/noty/ui"
)

func ParseUsers(usernames []string) []notion.NotionUser {
	userList := Users()
	var users = make([]notion.NotionUser, 0)
	for _, uf := range usernames {
		var found *notion.NotionUser = nil
		for _, u := range userList {
			if strings.Contains(strings.ToLower(u.Name), strings.ToLower(uf)) {
				found = &u
				break
			}
		}
		if found == nil {
			ui.PrintlnfWarn("no user found for '%s'", uf)
		} else {
			users = append(users, *found)
		}
	}
	return users
}
