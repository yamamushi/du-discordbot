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
	BankConfig		bankConfig 		`toml:"bank"`
	CasinoConfig	casinoConfig 	`toml:"casino"`
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
	PerPageCount int  `toml:"per_page_count"`
	Profiler bool `toml:"enable_profiler"`

}

type bankConfig struct {
	BankName 	string	`toml:"bank_name"`
	BankURL		string	`toml:"bank_url"`
	BankIconURL	string	`toml:"bank_icon_url"`
	Pin 		string 	`toml:"bank_pin"`
	Reset	bool	`toml:"reset_bank"`
	SeedWallet	int  `toml:"starting_bank_wallet_value"`
	SeedUserAccountBalance	int  `toml:"starting_user_account_value"`
	SeedUserWalletBalance	int  `toml:"starting_user_wallet_value"`
	BankMenuSlogan	string `toml:"bank_menu_slogan"`

}


type casinoConfig struct {
	CasinoName 	string	`toml:"casino_name"`
	Pin 		string 	`toml:"casino_pin"`
	Reset	bool	`toml:"reset_casino"`
	SeedWallet	int  `toml:"starting_casino_wallet_value"`
}



func ReadConfig(path string) ( config Config, err error) {

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err)
		return conf, err
	}

	return conf, nil
}
