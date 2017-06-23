package main

import (
	"github.com/BurntSushi/toml"
	"fmt"
)

type mainConfig struct {
	DiscordConfig   discordConfig 	`toml:"discord"`
	DBConfig		databaseConfig 	`toml:"database"`
	DUBotConfig 	dubotConfig 	`toml:"du-bot"`
}

type discordConfig struct {
	Token 	string 	`toml:"bot_token"`
}

type databaseConfig struct {
	Server  string  `toml:"hostname"`
	Port   	int 	`toml:"port"`
	User 	string 	`toml:"user"`
	Pass 	string 	`toml:"password"`
}

type dubotConfig struct {

	// Command Prefix
	CP 		string 	`toml:"command_prefix"`
	Playing string 	`toml:"default_now_playing"`
}

func ReadConfig(path string) ( config mainConfig, err error) {

	var conf mainConfig
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err)
		return conf, err
	}

	return conf, nil
}