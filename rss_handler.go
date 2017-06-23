package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

type RSSHandler struct {
	feeds	RSSFeeds
	db		*DBHandler
	conf	*mainConfig
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

	if m.Content == cp + "rss" {


		u, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			fmt.Println("error retrieving user:" + m.Author.ID)
		}
		println(u.ID)
		println(u.Admin)

		if u.Admin {
			s.ChannelMessageSend(m.ChannelID, "success" )
		}

	}

}