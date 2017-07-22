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

	message := strings.ToLower(m.Content)

	if strings.Contains(message, "lord jc"){
		s.ChannelMessageSend(m.ChannelID, "Yes, my child? \n http://i.imgur.com/DYq8TNe.jpg")
		return
	}

}
