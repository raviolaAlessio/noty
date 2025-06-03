package config

import (
	"fmt"
	"os"
	"path"

	"github.com/ravvio/noty/notion"
	"github.com/spf13/viper"
)

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(base, "noty"), nil
}

func Init() error {
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
			viper.WriteConfigAs(filepath)
			return filepath, nil
		} else {
			return "", err
		}
	}
	return viper.ConfigFileUsed(), nil
}

func TasksDatabaseID() string {
	return viper.GetString("tasks_database_id")
}

func ProjectsDatabaseID() string {
	return viper.GetString("projects_database_id")
}

func Users() []notion.NotionUser {
	users := viper.Get("users").([]any)
	res := make([]notion.NotionUser, 0, len(users))
	for _, user := range users {
		m := user.(map[string]any)
		res = append(res, notion.NotionUser{
			ID: m["id"].(string),
			Name: m["name"].(string),
		})
	}
	return res
}

func Projects() []notion.Project {
	projects := viper.Get("projects").([]any)
	res := make([]notion.Project, 0, len(projects))
	for _, project := range projects {
		m := project.(map[string]any)
		res = append(res, notion.Project{
			ID: m["id"].(string),
			Name: m["title"].(string),
		})
	}
	return res
}

func ProjectsMap() map[string]string {
	projects := viper.Get("projects").([]any)
	res := make(map[string]string)
	for _, project := range projects {
		m := project.(map[string]any)
		res[m["id"].(string)] = m["name"].(string)
	}
	return res
}
