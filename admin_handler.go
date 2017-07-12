package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

type AdminHandler struct {
	conf *Config
	db   *DBHandler
}

func (h *AdminHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	db := h.db.rawdb.From("Users")

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.Fields(m.Content)

	var u User
	db.One("ID", m.Author.ID, &u)
	if u.Admin != true {
		// We don't want to go any further
		return
	}

	// grant <user> role
	if message[0] == cp+"grant" && len(message) < 4 {
		s.ChannelMessageSend(m.ChannelID, "Expected 3 arguments: grant <user> <role>")
		return
	}

	if message[0] == cp+"grant" && len(message) > 3 {

	}
}
