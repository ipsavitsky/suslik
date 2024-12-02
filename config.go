package main

import (
	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
)

type Config struct {
	Token           string
	BaseURL         string
	ReviewerFileRef string
}

func parseConfig(filename string) Config {
	var conf Config
	_, err := toml.DecodeFile(filename, &conf)
	if err != nil {
		log.Fatalf("Error when reading configuration: %v", err)
	}

	if conf.Token == "" {
		log.Fatal("Empty GitLab token")
	}

	if conf.BaseURL == "" {
		log.Warn("Empty base url, setting default")
		conf.BaseURL = "https://gitlab.com/api/v4"
	}

	if conf.ReviewerFileRef == "" {
		log.Warn("Empty reviewer file ref, setting default")
		conf.ReviewerFileRef = "main"
	}

	return conf
}
