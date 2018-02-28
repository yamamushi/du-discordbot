package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// MemeHandler struct -> This operates without relying on a command string.
// This is supposed to read all messages and look for keywords to trigger a response.
type MemeHandler struct {
}

// Read function
func (h *MemeHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.ToLower(m.Content)

	if strings.Contains(message, "lord jc") {
		s.ChannelMessageSend(m.ChannelID, "Yes, my child? \n http://i.imgur.com/DYq8TNe.jpg")
		return
	}
	if strings.Contains(message, "jc") {
		s.ChannelMessageSend(m.ChannelID, "<:jc:236526936573214720>")
		return
	}
	if strings.Contains(message, "nyzaltar") {
		s.ChannelMessageSend(m.ChannelID, "<:nyzaltar:236531709632315402>")
		return
	}
}
