package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"os/user"
)

type PermissionsHandler struct {

	db *DBHandler
	conf *Config
	dg *discordgo.Session
	callback *CallbackHandler
	user *UserHandler

}


func (h *PermissionsHandler) Read() (s *discordgo.Session, m *discordgo.MessageCreate){

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	// Verify the user account exists (creates one if it doesn't exist already)
	h.user.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println("Error finding user")
		return
	}

	cp := h.conf.DUBotConfig.CP
	command := strings.Fields(m.Content)
	// We use this a bit, this is the author id formatted as a mention
	authormention := m.Author.Mention()
	mentions := m.Mentions


	if user.Owner{

	}
	if user.Admin {

		// Only admins have the permission to change the permission level of another command
		// registerCommand
	}
	if user.SModerator {

	}
	if user.JModerator {

	}
	if user.Editor {

	}
	if user.Recruiter {

	}
	if user.Streamer {

	}
	if user.Agora {

	}
	if user.Citizen {

		// ALL users have the citizen role. It is a permanent role once they have been registered in the database.

	}
}