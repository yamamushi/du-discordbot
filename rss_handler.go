package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

type RSSHandler struct {
	feeds	RSSFeeds
	db		*DBHandler
	conf	*mainConfig
	callback *CallbackHandler
	dg		*discordgo.Session
}

type RSSFeed struct {
	Name 	string
	URL   string
}

type RSSFeeds struct {
	Feeds[] RSSFeed
}


func (h *RSSHandler) addfeed() error {

	return nil

}

func (h *RSSHandler) menu(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if strings.HasPrefix(m.Content, cp + "rss") {

		command := strings.Fields(m.Content)

		// Grab our sender ID to verify if this user has permission to use this command
		u, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			fmt.Println("error retrieving user:" + m.Author.ID)
		}


		if u.Admin {

			if len(command) < 2{
				s.ChannelMessageSend(m.ChannelID, "Expected flag for 'rss' command" )
				return
			}

			if command[1] == "add" && len(command) > 2 {
				s.ChannelMessageSend(m.ChannelID, "Adding RSS Feed: " + command[2] + "Confirm? (Y/N)")
				message := m.Author.ID + " " + m.ChannelID + command[2]
				h.callback.Watch(m.Author.ID, m.ChannelID, h.ConfirmRSS, message, s, m)
				return
			}

			if command[1] == "add" && len(command) < 3 {
				s.ChannelMessageSend(m.ChannelID, "Insufficient arguments supplied")
			}
		}
	}
}


// This function ended up being unnecessary, just too sleepy to realize it at the time
/*
func (h *RSSHandler) AddRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

	commandlist := strings.Fields(command)

	if len(commandlist) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Invalid command received")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Adding: " + commandlist[2] + "Confirm? (Y/N)")

	message := m.Author.ID + " " + m.ChannelID
	h.callback.Watch(commandlist[0], commandlist[1], h.ConfirmRSS, message ,s, m)

}
*/


func (h *RSSHandler) ConfirmRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

	commandlist := strings.Fields(command)

	if len(commandlist) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Invalid command received")
		return
	}

	if m.Content == "Y" || m.Content == "y" {
		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "RSS Add Cancelled")

}