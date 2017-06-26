package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

type RSSHandler struct {
	db		*DBHandler
	conf	*Config
	callback *CallbackHandler
	dg		*discordgo.Session
}


func (h *RSSHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if strings.HasPrefix(m.Content, cp + "rss") {

		command := strings.Fields(m.Content)

		// Grab our sender ID to verify if this user has permission to use this command
		db := h.db.rawdb.From("Users")
		var user User
		err := db.One("ID", m.Author.ID, &user)
		if err != nil {
			fmt.Println("error retrieving user:" + m.Author.ID)
		}


		if user.Admin {

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

			if command[1] == "list" && len(command) == 2 {
				formatted := h.GetRSSList(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, formatted)
				return
			}

			if command[1] == "get" && len(command) == 2 {
				s.ChannelMessageSend(m.ChannelID, "Please supply a feed URL: ")
				return
			}

			if command[1] == "get" && len(command) > 2 {
				message, err := h.GetLatestItem(m.ChannelID, command[2])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, message)
				return
			}

		}
	}
}

func (h *RSSHandler) GetLatestItem(channel string, url string) (formatted string, err error) {
	rss := RSS{db: h.db}

	item, err := rss.GetLatestItem(url, channel)

	if err != nil {
		return formatted, err
	}

	if item.Twitter {
		formatted = "Latest Tweet from " + item.Author + "\n\n"
	} else if item.Reddit {

		var subreddit string
		if strings.HasPrefix(url, "https://www.reddit.com"){
			subreddit = strings.TrimPrefix(strings.TrimSuffix(url, ".rss"), "https://www.reddit.com")
		} else if strings.HasPrefix(url, "http://www.reddit.com"){
			subreddit = strings.TrimPrefix(strings.TrimSuffix(url, ".rss"), "http://www.reddit.com")
		}
		//subreddit = strings.Trim(subreddit, "http://www.reddit.com")
		formatted = "Latest Update from " + subreddit + " - " + "https://www.reddit.com/user"+item.Author + " \n\n"
		formatted = formatted + item.Title + "\n\n"
		//formatted = formatted + item.Content + "\n"
		//formatted = formatted + item.Description + "\n"
		formatted = formatted + item.Link + "\n"
		formatted = formatted + item.Published + "\n"

	} else {
		formatted = "Latest update from " + url + "\n"
		formatted = formatted + item.Author + "\n"
		formatted = formatted + item.Title + "\n"
		formatted = formatted + item.Content + "\n"
		//formatted = formatted + item.Description + "\n"
		formatted = formatted + item.Link + "\n"
		formatted = formatted + item.Published + "\n"
	}

	return formatted, nil
}


func (h *RSSHandler) GetRSSList(channel string) (formatted string) {

	rss := RSS{db: h.db}
	feeds, err := rss.GetChannel(channel)
	if err != nil {
		return "No subscriptions could be found for this channel"
	}

	formatted = "RSS Subscriptions for this Channel:" + "\n"
	for _, i := range feeds {
		formatted = formatted + i.Title + " - " + i.URL + "\n"
	}

	//formatted = formatted + "```"

	return formatted

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



func (h *RSSHandler) ConfirmRSS(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp){
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	if m.Content == "Y" || m.Content == "y" {

		rss := RSS{db: h.db}
		err := rss.Subscribe(url, "", m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error Subscribing to URL: " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: " + url )
		return
	}

	s.ChannelMessageSend(m.ChannelID, "RSS Add Cancelled")
}

