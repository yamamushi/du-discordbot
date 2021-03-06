package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"time"
)

// Config struct
type Config struct {
	DiscordConfig discordConfig     `toml:"discord"`
	DBConfig      databaseConfig    `toml:"database"`
	DUBotConfig   dubotConfig       `toml:"du-bot"`
	BankConfig    bankConfig        `toml:"bank"`
	CasinoConfig  casinoConfig      `toml:"casino"`
	RolesConfig   rolesConfig       `toml:"roles"`
	APIConfig     apiConfig         `toml:"api"`
	Recruitment   recruitmentConfig `toml:"recruitment"`
	Reactions     ReactionsConfig   `toml:"reactions"`
	Rabbit        RabbitConfig      `toml:"rabbit"`
}

// discordConfig struct
type discordConfig struct {
	Token   string `toml:"bot_token"`
	AdminID string `toml:"admin_id"`
	GuildID string `toml:"guild_id"`
	DevID   string `toml:"dev_id"`
}

// databaseConfig struct
type databaseConfig struct {
	DBFile             string `toml:"filename"`
	MongoHost          string `toml:"mongohost"`
	MongoDB            string `toml:"mongodb"`
	BackerRecordColumn string `toml:"backerrecordcolumn"`
	InfoRecordColumn   string `toml:"inforecordcolumn"`
	MongoUser          string `toml:"mongouser"`
	MongoPass          string `toml:"mongopass"`
}

// dubotConfig struct
type dubotConfig struct {

	// Command Prefix
	CP               string        `toml:"command_prefix"`
	Playing          string        `toml:"default_now_playing"`
	RSSTimeout       time.Duration `toml:"rss_fetch_timeout"`
	StatsTimeout     time.Duration `toml:"stats_fetch_timeout"`
	Notifications    time.Duration `toml:"notifications_update_timeout"`
	GiveawayTimer    time.Duration `toml:"giveaway_timer"`
	GiveawayChannel  string        `toml:"giveaway_channel"`
	PerPageCount     int           `toml:"per_page_count"`
	LuaTimeout       int           `toml:"lua_timeout"`
	Profiler         bool          `toml:"enable_profiler"`
	MaxAudioDuration int           `toml:"max_audio_duration"`
}

// bankConfig struct
type bankConfig struct {
	BankName               string `toml:"bank_name"`
	BankURL                string `toml:"bank_url"`
	BankIconURL            string `toml:"bank_icon_url"`
	Pin                    string `toml:"bank_pin"`
	Reset                  bool   `toml:"reset_bank"`
	SeedWallet             int    `toml:"starting_bank_wallet_value"`
	SeedUserAccountBalance int    `toml:"starting_user_account_value"`
	SeedUserWalletBalance  int    `toml:"starting_user_wallet_value"`
	BankMenuSlogan         string `toml:"bank_menu_slogan"`
}

// casinoConfig struct
type casinoConfig struct {
	CasinoName string `toml:"casino_name"`
	Pin        string `toml:"casino_pin"`
	Reset      bool   `toml:"reset_casino"`
	SeedWallet int    `toml:"starting_casino_wallet_value"`
}

// rolesConfig struct
type rolesConfig struct {
	RoleTimer               time.Duration `toml:"roles_timer"`
	RoleUpdaterTimer        time.Duration `toml:"roles_updater_timer"`
	PatronRoleID            string        `toml:"patron_role"`
	SponsorRoleID           string        `toml:"sponsor_role"`
	ContributorRoleID       string        `toml:"contributor_role"`
	ATVRoleID               string        `toml:"atv_role"`
	IronRoleID              string        `toml:"iron_role"`
	BronzeRoleID            string        `toml:"bronze_role"`
	SilverRoleID            string        `toml:"silver_role"`
	GoldRoleID              string        `toml:"gold_role"`
	SapphireRoleID          string        `toml:"sapphire_role"`
	RubyRoleID              string        `toml:"ruby_role"`
	EmeraldRoleID           string        `toml:"emerald_role"`
	DiamondRoleID           string        `toml:"diamond_role"`
	KyriumRoleID            string        `toml:"kyrium_role"`
	ForumLinkedRoleID       string        `toml:"forumlinked_role"`
	ATVForumLinkedRoleID    string        `toml:"atvforumlinked_role"`
	PreAlphaForumLinkedRole string        `toml:"prealpha_role"`
	KickstarterRole         string        `toml:"kickstarter_role"`
	AlphaAuthorizedRole     string        `toml:"alpha_role"`

	NDAChannelID string `toml:"nda_channel_id"`
}

type apiConfig struct {
	Strawpoll      string `toml:"strawpoll_api"`
	WordnikKey     string `toml:"wordnick_key"`
	OpenWeatherKey string `toml:"openweather_key"`
}

type recruitmentConfig struct {
	RecruitmentChannel       string        `toml:"recruitment_channel"`
	RecruitmentTimeout       time.Duration `toml:"recruitment_timeout"`
	RecruitmentExpiration    time.Duration `toml:"recruitment_expiration_timeout"`
	RecruitmentReminder      time.Duration `toml:"recruitment_reminder_timeout"`
	RecruitmentShuffleCount  int           `toml:"recruitment_shuffle_count"`
	RecruitmentWaitOnStartup bool          `toml:"recruitment_wait_on_startup"`
}

type ReactionsConfig struct {
	ReactionsExpiration     time.Duration `toml:"reactions_timeout"`
	InfoReactionsExpiration time.Duration `toml:"info_reactions_timeout"`
}

type RabbitConfig struct {
	RabbitRandomWeight int           `toml:"rabbit_random"`
	RabbitCount        int           `toml:"rabbit_count"`
	RabbitTimer        time.Duration `toml:"rabbit_timer"`
	RabbitExpiration   time.Duration `toml:"rabbit_expiration"`
}

// ReadConfig function
func ReadConfig(path string) (config Config, err error) {

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err)
		return conf, err
	}

	return conf, nil
}
