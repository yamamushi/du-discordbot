package main

/*

Our interface to our command registry

 */


import (
	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	callback *CallbackHandler
	conf *Config
	db *DBHandler
	perm *PermissionsHandler
	registry *CommandRegistry
	dg *discordgo.Session
	user *UserHandler
}


func (h *CommandHandler) Init(){

	// Setup our command registry interface
	h.registry = new(CommandRegistry)
	h.registry.db = h.db

}




func (h *CommandHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Check for safety
	if !SafeInput(s, m){
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		return
	}
	if !user.CheckRole("admin"){
		return
	}

	command, message := CleanCommand(m.Content, h.conf)

	if command == "command" {
		h.ReadCommand(message, s, m)

	}
}


func (h *CommandHandler) ReadCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate){

	if len(message) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<command> requires an argument")
		return
	}

	if message[0] == "enable"{
		if len(message) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<enable> requires at least one argument")
			return
		}
		h.EnableCommand(message, s, m)
		return
	}
	if message[0] == "disable"{
		if len(message) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<disable> requires at least one argument")
			return
		}
		h.DisableCommand(message, s, m)
		return

	}
	if message[0] == "usage"{
		if len(message) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<usage> requires at least one argument")
			return
		}
		h.DisplayUsage(message, s, m)
		return

	}
	if message[0] == "description"{
		if len(message) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<description> requires at least one argument")
			return
		}
		h.DisplayDescription(message, s, m)
		return

	}
	if message[0] == "list" {
		h.ListCommands(s, m)
	}
}

func (h *CommandHandler) EnableCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	h.registry.AddChannel(message[1], m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, "Command " + message[1] + " enabled for this channel")
	return
}

func (h *CommandHandler) DisableCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	h.registry.RemoveChannel(message[1], m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, "Command " + message[1] + " disabled for this channel")
	return
}

func (h *CommandHandler) DisplayUsage(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, err := h.registry.GetCommand(message[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	formattedmessage := ":\n" + " Usage guide for " + message[1] + "\n"
	formattedmessage = formattedmessage + "```"
	formattedmessage = formattedmessage + command.Usage
	formattedmessage = formattedmessage + "```"

	s.ChannelMessageSend(m.ChannelID, formattedmessage)
	return
}

func (h *CommandHandler) DisplayDescription(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, err := h.registry.GetCommand(message[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	formattedmessage := ":\n" + " Description for " + message[1] + "\n"
	formattedmessage = formattedmessage + "```"
	formattedmessage = formattedmessage + command.Description
	formattedmessage = formattedmessage + "```"

	s.ChannelMessageSend(m.ChannelID, formattedmessage)
	return
}


func (h *CommandHandler) ListCommands(s *discordgo.Session, m *discordgo.MessageCreate) {

	recordlist, err := h.registry.CommandsForChannel(m.ChannelID)
	if err != nil{

		if err.Error() == "not found" {
			s.ChannelMessageSend(m.ChannelID, "No commands for this channel found")
			return
		}
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	formattedmessage := ":\n" + " Command list for this channel "
	formattedmessage = formattedmessage + "\n-------------------------------\n"
	formattedmessage = formattedmessage + "```"

	for _, command := range recordlist {
		formattedmessage = formattedmessage + command.Command + " " + command.Description + "\n"
	}
	formattedmessage = formattedmessage + "```"

	s.ChannelMessageSend(m.ChannelID, formattedmessage)
	return
}


