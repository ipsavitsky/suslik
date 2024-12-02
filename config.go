package main

import (
	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
)

type Config struct {
	Token string
	ReviewerFileRef string
}

func parseConfig(filename string) Config {
	var conf Config
	_, err := toml.DecodeFile(filename, &conf)
	if err != nil {
		log.Fatalf("Error when reading configuration: %v", err)
	}
	return conf
}
