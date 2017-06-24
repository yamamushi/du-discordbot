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

			if command[1] == "add" && len(command) == 2 {
				s.ChannelMessageSend(m.ChannelID, "Please supply a feed URL: ")
				message := ""
				h.callback.Watch( h.GetRSS, GetUUID(), message, s, m)
				return
			}

			if command[1] == "add" && len(command) > 2 {
				s.ChannelMessageSend(m.ChannelID, "Add RSS Feed: " + command[2] + " Confirm? (Y/N)")
				message := command[2]
				h.callback.Watch( h.ConfirmRSS, GetUUID(), message, s, m)
				return
			}

		}
	}
}



func (h *RSSHandler) GetRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// In this handler we don't do anything with the command string, instead we grab the response from m.Content

	// We do this to avoid having duplicate commands overrunning each other
	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp){
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	// A poor way of checking the validity of the RSS url for now
	if m.Content == "" {
		s.ChannelMessageSend(m.ChannelID, "Invalid Command Received")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Adding: " + m.Content + " Confirm? (Y/N)")

	h.callback.Watch( h.ConfirmRSS, GetUUID(), m.Content, s, m)

}



func (h *RSSHandler) ConfirmRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp){
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	if m.Content == "Y" || m.Content == "y" {
		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: " + command )
		h.AddRSS(command)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "RSS Add Cancelled")
}


func (h *RSSHandler) AddRSS(url string) error {
	fmt.Println("Adding RSS Feed: " + url)
	return nil
}
