package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Telegram TelegramConfig `toml:"telegram"`
	GitLab   GitLabConfig   `toml:"gitlab"`
}

type GitLabConfig struct {
	Token  string   `toml:"token"`
	URL    string   `toml:"url"`
	Groups []string `toml:"groups"`
}

type TelegramConfig struct {
	Token   string `toml:"token"`
	Channel int64  `toml:"channel"`
}

var config Config

func init() {
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		panic(err)
	}
}
