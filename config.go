package main

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
)

type ModeType string

const (
	CI         ModeType = "ci"
	Standalone ModeType = "standalone"
)

type Config struct {
	Token           string
	BaseURL         string
	ReviewerFileRef string
	Mode            ModeType
	PollDelay       time.Duration
}

func parseConfig(filename string, mode string) Config {
	var conf Config
	_, err := toml.DecodeFile(filename, &conf)
	if err != nil {
		log.Fatalf("Error when reading configuration: %v", err)
	}

	if conf.Token == "" {
		log.Warn("No token in configuration, reading environment")
		token, found := os.LookupEnv("SUSLIK_GITLAB_TOKEN")
		if !found {
			log.Fatal("Empty GitLab token")
		}
		conf.Token = token
	}

	if conf.BaseURL == "" {
		log.Warn("Empty base url, setting default")
		conf.BaseURL = "https://gitlab.com/api/v4"
	}

	if conf.ReviewerFileRef == "" {
		log.Warn("Empty reviewer file ref, setting default")
		conf.ReviewerFileRef = "main"
	}

	if conf.PollDelay == 0 {
		log.Warn("Empty poll duration, setting default")
		conf.PollDelay = 10 * time.Second
	}

	if mode == string(Standalone) {
		conf.Mode = Standalone
	} else if mode == string(CI) {
		conf.Mode = CI
	} else {
		log.Fatal("Unknown mode: %s", mode)
	}

	return conf
}
