package main

/*

Our interface to our command registry

 */


import (
	"strings"

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

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		return
	}

	if !user.CheckRole("admin"){
		return
	}

	// Set our command prefix to the default one within our config file
	cp := h.conf.DUBotConfig.CP
	message := strings.Fields(m.Content)
	if len(message) < 1 {
		return
	}

	message[0] = strings.Trim(message[0], cp)
	command := message[0]

	// If the message read is "hello" reply in the same channel with "Hi!"
	if command == "command" {
		if len(message) < 2 {
			s.ChannelMessageSend( m.ChannelID, "<command> expects an argument")
			return
		}
	}
	if command == "prompt" {

		s.ChannelMessageSend(m.ChannelID, "Y or N?")

		// CallbackHandler.Watch expects a message, which it could
		// Also use to pass in arguments directly to a callback
		// That needs some setup or other parameters
		// In this example, our sub callback doesn't do anything
		// With the message it receives, so we can leave it blank
		message := ""

		// (callback, uuid, message, session, messagecreate)
		h.callback.Watch( h.Validate, GetUUID(), message, s, m)
	}
}


// All sub-callbacks MUST have this function signature
// func( string, *discordgo.Session, *discordgo.MessageCreate)
func (h *CommandHandler) Validate(message string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Setup our command prefix
	cp := h.conf.DUBotConfig.CP
	// We want to cancel our command if another one is called by our user
	// We do this to avoid having duplicate/similar commands overrunning each other
	if strings.HasPrefix(m.Content, cp){
		s.ChannelMessageSend(m.ChannelID, "Prompt Cancelled")
		return
	}

	// Check
	if m.Content == "Y" || m.Content == "y" {
		s.ChannelMessageSend(m.ChannelID, "You Selected Yes" )
		return
	}
	if m.Content == "N" || m.Content == "n" {
		s.ChannelMessageSend(m.ChannelID, "You Selected No" )
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Invalid Response")

}