package main

import (
	"flag"

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

	var confFile string
	var mode string

	flag.StringVar(&confFile, "c", "conf.toml", "Path to configuration file")
	flag.StringVar(&mode, "m", "standalone", "suslik mode (ci or standalone)")
	flag.Parse()

	if (mode != "standalone") && (mode != "ci") {
		log.Fatalf("Unknown mode: %s", mode)
	}

	conf := parseConfig(confFile)

	app := app{
		client: getGitlabClient(conf.Token, conf.BaseURL),
		conf:   conf,
	}

	if mode == "standalone" {
		app.loop()
	} else {
		app.run()
	}
}
