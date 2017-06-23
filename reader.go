package main

import (
	"github.com/bwmarrin/discordgo"
)

type MessageReader struct {

	db *DBHandler
	conf *mainConfig

}


// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (r *MessageReader) read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := r.conf.DUBotConfig.CP

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == cp + "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == cp + "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

}

