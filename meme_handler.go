package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"

	"strconv"
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
	//fmt.Print(message)

	if strings.Contains(message, "lord jc") {
		s.ChannelMessageSend(m.ChannelID, "Yes, my child? \n http://i.imgur.com/DYq8TNe.jpg")
		return
	}
	if strings.Contains(message, "scp-"){
		payload := strings.Split(message, " ")
		for _, word := range payload {
			if strings.HasPrefix(word, "scp-") {
				scpnum := strings.TrimPrefix(word, "scp-")
				val, err := strconv.Atoi(scpnum)
				if err == nil && val > 0 {
					s.ChannelMessageSend(m.ChannelID, "http://www.scp-wiki.net/"+word)
				}
			}
		}
	}
	if MessageHasMeme(message, "jc") {
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, ":jc:236526936573214720")
	}
	if MessageHasMeme(message, "nyzaltar") {
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, ":nyzaltar:236531709632315402")
	}
	if MessageHasMeme(message, "oli") {
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, ":oli:421481974201319424")
	}
	if MessageHasMeme(message, "vape") {
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, ":vapenation:360989703215775754")
	}
	if MessageHasMeme(message, "diesel") {
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, ":diesel:442400293972475904")
	}
	/*
	if MessageHasMeme(message, "thanks") || MessageHasMeme(message, "thank you") ||
		MessageHasMeme(message, "danke") || MessageHasMeme(message, "gracias") ||
			MessageHasMeme(message, "tom hanks"){
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, ":thanks:297438919165739019")
	}
	*/
	return
}
