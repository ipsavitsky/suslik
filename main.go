package main

import (
	"flag"

	"github.com/charmbracelet/log"

	gitlab "gitlab.com/gitlab-org/api/client-go"
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

	var confFile string
	var mode string

	flag.StringVar(&confFile, "c", "conf.toml", "Path to configuration file")
	flag.StringVar(&mode, "m", string(Standalone), "suslik mode (ci or standalone)")
	flag.Parse()

	conf := parseConfig(confFile, mode)

	app := app{
		client: getGitlabClient(conf.Token, conf.BaseURL),
		conf:   conf,
	}

	if conf.Mode == Standalone {
		app.loop()
	} else {
		app.run()
	}
}
