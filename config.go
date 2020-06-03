package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Telegram TelegramConfig `toml:"telegram"`
	GitLab   GitLabConfig   `toml:"gitlab"`
}

type GitLabConfig struct {
	Token string `toml:"token"`
	URL   string `toml:"url"`
}

type ChannelConfig struct {
	ID       int64    `toml:"id"`
	Groups   []string `toml:"groups"`
	Projects []string `toml:"projects"`
}

type TelegramConfig struct {
	Token    string          `toml:"token"`
	Channels []ChannelConfig `toml:"channels"`
}

var config Config
var allGroups []string

func init() {
	_, err := toml.DecodeFile("config.toml", &config)
	for _, channel := range config.Telegram.Channels {
		for _, channelGroup := range channel.Groups {
			allGroups = append(allGroups, channelGroup)
		}
	}
	if err != nil {
		panic(err)
	}
}
