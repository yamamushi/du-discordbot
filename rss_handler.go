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

			if command[1] == "add" {
				s.ChannelMessageSend(m.ChannelID, "Test Success, testing queue")
				h.callback.Watch(m.Author.ID, m.ChannelID)
				return
			}

		}

	}

}