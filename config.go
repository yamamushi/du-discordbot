package main

import (
	"github.com/BurntSushi/toml"
	"fmt"
	"time"
)

type Config struct {
	DiscordConfig	discordConfig 	`toml:"discord"`
	DBConfig		databaseConfig 	`toml:"database"`
	DUBotConfig		dubotConfig 	`toml:"du-bot"`
}

type discordConfig struct {
	Token 	string 	`toml:"bot_token"`
	AdminID	string	`toml:"admin_id"`
}

type databaseConfig struct {
	DBFile  string  `toml:"filename"`
}

type dubotConfig struct {

	// Command Prefix
	CP 		string 	`toml:"command_prefix"`
	Playing string 	`toml:"default_now_playing"`
	RSSTimeout time.Duration  `toml:"rss_fetch_timeout"`
}

func ReadConfig(path string) ( config Config, err error) {

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err)
		return conf, err
	}

	return conf, nil
}