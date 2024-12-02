package main

import (
	"github.com/charmbracelet/log"

	gitlab "github.com/xanzy/go-gitlab"
)

type app struct {
	client *gitlab.Client
	conf   Config
}

func getGitlabClient(token string, baseURL string) *gitlab.Client {
	git, err := gitlab.NewClient(
		token,
		gitlab.WithBaseURL(baseURL),
	)
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
	}
	return git
}

func (a *app) getCurrentUser() *gitlab.User {
	user, _, err := a.client.Users.CurrentUser()
	if err != nil {
		log.Errorf("Failed to get current user: %v", err)
	}
	log.Debugf("Username: %s", user.Name)
	return user
}

func main() {
	log.SetLevel(log.DebugLevel)

	conf := parseConfig("conf.toml")

	app := app{
		client: getGitlabClient(conf.Token, conf.BaseURL),
		conf: conf,
	}

	app.loop()
}
