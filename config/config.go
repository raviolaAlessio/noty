package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ravvio/noty/notion"
	"github.com/spf13/viper"
)

const (
	KeyTasksDatabaseID    = "tasks_database_id"
	KeyProjectsDatabaseID = "projects_database_id"
	KeySprintsDatabaseID  = "sprints_database_id"
	KeyHoursDatabaseID    = "hours_database_id"
	KeyUsers              = "users"
	KeyProjects           = "projects"
	KeyUseEmotes          = "use_emotes"
	KeyStatusEmotes       = "status_emotes"
	KeyDatetimeFormat     = "datetime_format"
	KeyDateFormat         = "datetime_format"
)

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(base, "noty"), nil
}

func Init() error {
	viper.SetDefault(KeyTasksDatabaseID, "")
	viper.SetDefault(KeyProjectsDatabaseID, "")
	viper.SetDefault(KeySprintsDatabaseID, "")
	viper.SetDefault(KeyHoursDatabaseID, "")

	viper.SetDefault(KeyUseEmotes, true)
	viper.SetDefault(KeyStatusEmotes, map[string]string{
		"not_started":  "❄️",
		"in_progress":  "🚀",
		"to_be_tested": "💣",
		"in_testing":   "💥",
		"done":         "✅",
		"cancelled":    "❌",
	})

	viper.SetDefault(KeyDatetimeFormat, "2006-01-02 15:04")
	viper.SetDefault(KeyDateFormat, "2006-01-02")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	viper.AddConfigPath(dir)
	viper.AddConfigPath("$HOME/.config/noty")
	viper.AddConfigPath("$HOME/.noty")
	return nil
}

func Load() (bool, error) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return false, nil
		} else {
			return false, fmt.Errorf("error parsing configuration: %s", err)
		}
	}
	return true, nil
}

func Save() (string, error) {
	if err := viper.WriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			dir, err := ConfigDir()
			if err != nil {
				return "", err
			}
			filepath := path.Join(dir, "config.yaml")
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				return "", err
			}
			err = viper.WriteConfigAs(filepath)
			if err != nil {
				return "", nil
			}
			return filepath, nil
		} else {
			return "", err
		}
	}
	return viper.ConfigFileUsed(), nil
}

func TasksDatabaseID() string {
	return viper.GetString(KeyTasksDatabaseID)
}

func ProjectsDatabaseID() string {
	return viper.GetString(KeyProjectsDatabaseID)
}

func SprintsDatabaseID() string {
	return viper.GetString(KeySprintsDatabaseID)
}

func HoursDatabaseID() string {
	return viper.GetString(KeyHoursDatabaseID)
}

func UseEmotes() bool {
	return viper.GetBool(KeyUseEmotes)
}

func StatusEmotes() map[string]string {
	return viper.GetStringMapString(KeyStatusEmotes)
}

func StatusEmote(value string) string {
	return StatusEmotes()[strings.ReplaceAll(strings.ToLower(value), " ", "_")]
}

func DatetimeFormat() string {
	return viper.GetString(KeyDatetimeFormat)
}

func DateFormat() string {
	return viper.GetString(KeyDateFormat)
}

func Users() []notion.NotionUser {
	users := viper.Get(KeyUsers).([]any)
	res := make([]notion.NotionUser, 0, len(users))
	for _, user := range users {
		m := user.(map[string]any)
		res = append(res, notion.NotionUser{
			ID:   m["id"].(string),
			Name: m["name"].(string),
		})
	}
	return res
}

func Projects() []notion.Project {
	projects := viper.Get(KeyProjects).([]any)
	res := make([]notion.Project, 0, len(projects))
	for _, project := range projects {
		m := project.(map[string]any)
		res = append(res, notion.Project{
			ID:   m["id"].(string),
			Name: m["name"].(string),
		})
	}
	return res
}

func ProjectsMap() map[string]string {
	projects := viper.Get(KeyProjects).([]any)
	res := make(map[string]string)
	for _, project := range projects {
		m := project.(map[string]any)
		res[m["id"].(string)] = m["name"].(string)
	}
	return res
}
