package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"time"
)

// NotificationsHandler struct
type RoleHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db *DBHandler
}

// Init function
func (h *RoleHandler) Init() {
	h.RegisterCommands()
}


// RegisterCommands function
func (h *RoleHandler) RegisterCommands() (err error) {
	h.registry.Register("roles", "Manage roles system", "list|edit|init|sync")
	return nil
}

// Read function
func (h *RoleHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"roles") {
		if h.registry.CheckPermission("roles", m.ChannelID, user) {

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

func (h *RoleHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command " + command + " expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "edit" {
		_, commandpayload := SplitPayload(payload)
		h.RoleEdit(commandpayload, s, m)
		return
	}
	if payload[0] == "init" {
		_, commandpayload := SplitPayload(payload)
		h.RoleInit(commandpayload, s, m)
		return
	}
	if payload[0] == "list" {
		_, commandpayload := SplitPayload(payload)
		h.RoleList(commandpayload, s, m)
		return
	}
	if payload[0] == "sync" {
		_, commandpayload := SplitPayload(payload)
		h.RoleSync(commandpayload, s, m)
		return
	}
}


// HelpOutput function
func (h *RoleHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate){
	output := "Command usage for giveaway: \n"
	output = output + "```\n"
	output = output + "list: list roles and their id's\n"
	output = output + "edit: adjust the timer for a role\n"
	output = output + "init: initialize the roles system\n"
	output = output + "sync: manually sync all users\n"
	output = output + "debug: provides debug output\n"
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}



func (h *RoleHandler) RoleSynchronizer(s *discordgo.Session) {
	for true {
		// Only run every X minutes
		time.Sleep(h.conf.RolesConfig.RoleTimer * time.Minute)

	}
}



func (h *RoleHandler) RoleEdit(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'edit' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *RoleHandler) RoleList(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'update' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *RoleHandler) RoleInit(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'init' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *RoleHandler) RoleSync(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'sync' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}