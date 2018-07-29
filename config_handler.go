package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"errors"
	"strconv"
	)

// ConfigHandler handles global config options that can be changed from within discord
type ConfigHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	configdb *ConfigDB
}


// Init function
func (h *ConfigHandler) Init() {
	h.RegisterCommands()
	h.configdb = &ConfigDB{db: h.db}
}

// RegisterCommands function
func (h *ConfigHandler) RegisterCommands() (err error) {

	h.registry.Register("config", "Manage config options for this server", "enable|disable|set")
	return nil

}

// Read function
func (h *ConfigHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"config") {
		if h.registry.CheckPermission("config", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Admin {
				h.ParseCommand(command, s, m)
			}
		}
	}

}



func (h *ConfigHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(command) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'config' command, see usage for more info")
		return
	}

	if command[1] == "enable" {
		if len(command) <= 2 {
			s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		} else {
			err := h.Enable(command[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Option " + command[2] + " enabled.")
			return
		}
		return
	}

	if command[1] == "disable" {
		if len(command) <= 2 {
			s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		} else {
			err := h.Disable(command[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Option " + command[2] + " disabled.")
			return
		}
	}

	if command[1] == "set" {
		if len(command) <= 3 {
			s.ChannelMessageSend(m.ChannelID, "Command expects two arguments")
		} else {
			option := command[2]
			value, err := strconv.Atoi(command[3])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid value: " + err.Error())
				return
			}
			err = h.Set(option, value)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Option " + command[2] + " set to " + command[3])
			return
		}
	}

	if command[1] == "write" {
		if len(command) <= 3 {
			s.ChannelMessageSend(m.ChannelID, "Command expects two arguments")
		} else {
			option := command[2]
			err := h.Write(option, command[3])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Wrote " + command[3] + " to " + command[2])
			return
		}
	}

	if command[1] == "get" {
		if len(command) <= 2 {
			s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		} else {
			h.Get(command[2], s, m)
			return
		}
	}
}

func (h *ConfigHandler) ValidateConfigName(configname string) bool {

	if configname == "autoland"{
		return true
	}
	if configname == "recruitment-timer"{
		return true
	}
	if configname == "recruitment-admin-channel"{
		return true
	}
	if configname == "recruitment-expiration"{
		return true
	}
	if configname == "recruitment-reminder"{
		return true
	}
	if configname == "reactions-expiration" {
		return true
	}
	if configname == "rabbit-timer" {
		return true
	}
	if configname == "rabbit-expiration" {
		return true
	}
	if configname == "rabbit-count" {
		return true
	}
	if configname == "rabbit-random" {
		return true
	}
	if configname == "rabbit-channel" {
		return true
	}

	return false
}

func (h *ConfigHandler) Enable(configname string) (err error) {

	if !h.ValidateConfigName(configname) {
		return errors.New("invalid config " + configname)
	}

	entry, err := h.configdb.GetConfigFromDB(configname)
	if err != nil {
		entry := ConfigEntry{Name:configname, Enabled:true}
		return h.configdb.AddConfigToDB(entry)
	}
	entry.Enabled = true
	return h.configdb.AddConfigToDB(entry)
}

func (h *ConfigHandler) Disable(configname string) (err error) {
	if !h.ValidateConfigName(configname) {
		return errors.New("invalid config " + configname)
	}

	entry, err := h.configdb.GetConfigFromDB(configname)
	if err != nil {
		entry := ConfigEntry{Name:configname, Enabled:false}
		return h.configdb.AddConfigToDB(entry)
	}
	entry.Enabled = false
	return h.configdb.AddConfigToDB(entry)
}

func (h *ConfigHandler) Set(configname string, value int) (err error) {
	if !h.ValidateConfigName(configname) {
		return errors.New("invalid config " + configname)
	}
	entry, err := h.configdb.GetConfigFromDB(configname)
	if err != nil {
		entry := ConfigEntry{Name:configname, Value:value}
		return h.configdb.AddConfigToDB(entry)
	}
	entry.Value = value
	return h.configdb.AddConfigToDB(entry)
}

func (h *ConfigHandler) Write(configname string, value string) (err error) {
	if !h.ValidateConfigName(configname) {
		return errors.New("invalid config " + configname)
	}
	entry, err := h.configdb.GetConfigFromDB(configname)
	if err != nil {
		entry := ConfigEntry{Name:configname, Setting:value}
		return h.configdb.AddConfigToDB(entry)
	}
	entry.Setting = value
	return h.configdb.AddConfigToDB(entry)
}

func (h *ConfigHandler) Get(configname string, s *discordgo.Session, m *discordgo.MessageCreate)  {

	if !h.ValidateConfigName(configname) {
		s.ChannelMessageSend(m.ChannelID, "Invalid option " + configname)
		return
	}

	entry, err := h.configdb.GetConfigFromDB(configname)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}
	output := "Config: \n```"
	output = output + "Name: " + entry.Name + "\n"
	output = output + "Enabled: " + strconv.FormatBool(entry.Enabled) + "\n"
	output = output + "Value: " + strconv.Itoa(entry.Value) + "\n"
	output = output + "```"

	s.ChannelMessageSend(m.ChannelID, output)
	return
}