package main

import (
	"github.com/charmbracelet/log"

	gitlab "github.com/xanzy/go-gitlab"
)

type app struct {
	client *gitlab.Client
	conf   Config
}

func get_gitlab_client(token string) *gitlab.Client {
	git, err := gitlab.NewClient(token)
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
	}
	return git
}

func (a *app) get_current_user() *gitlab.User {
	user, _, err := a.client.Users.CurrentUser()
	if err != nil {
		log.Errorf("Failed to get current user: %v", err)
	}
	log.Debugf("Name: %s", user.Name)
	return user
}

func main() {
	log.SetLevel(log.DebugLevel)

	conf := parseConfig("conf.toml")

	app := app{
		client: get_gitlab_client(conf.Token),
		conf: conf,
	}

	app.loop()
}
