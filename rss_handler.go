package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
	"time"
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
				h.callback.Watch( h.AddRSS, GetUUID(), message, s, m)
				return
			}

			if command[1] == "add" && len(command) > 2 {
				s.ChannelMessageSend(m.ChannelID, "Add RSS Feed: " + command[2] + " Confirm? (Y/N)")
				message := command[2]
				h.callback.Watch( h.ConfirmAddRSS, GetUUID(), message, s, m)
				return
			}

			if command[1] == "remove" && len(command) == 2 {
				s.ChannelMessageSend(m.ChannelID, "Please supply a feed URL: ")
				message := ""
				h.callback.Watch( h.RemoveRSS, GetUUID(), message, s, m)
				return
			}

			if command[1] == "remove" && len(command) > 2 {
				s.ChannelMessageSend(m.ChannelID, "Remove RSS Feed: " + command[2] + " Confirm? (Y/N)")
				message := command[2]
				h.callback.Watch( h.ConfirmRemoveRSS, GetUUID(), message, s, m)
				return
			}

			if command[1] == "list" && len(command) == 2 {
				formatted := h.GetRSSList(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, formatted)
				return
			}

			if command[1] == "get" && len(command) == 2 {
				h.GetAllLatest(s, m)
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

func (h *RSSHandler) UpdateRSSFeeds(s *discordgo.Session) {

	for true {
		// Only run every X minutes
		time.Sleep(h.conf.DUBotConfig.RSSTimeout * time.Minute)

		rss := RSS{db: h.db}
		rssfeeds, err := rss.GetDB()
		if err != nil {
			fmt.Println("Error Reading Database RSS! - " + err.Error())
			//return
			break
		}

		for _, feed := range rssfeeds {


			now := time.Now()
			lastrun := feed.LastRun
			expires := now.Sub(lastrun)

			if expires > h.conf.DUBotConfig.RSSTimeout * time.Minute {

				skip := false
				item, err := rss.GetLatestItem(feed.URL, feed.ChannelID)
				if err != nil {
					fmt.Println(err.Error())
					rss.Unsubscribe(feed.URL, feed.ChannelID) // Unsubscribe us if we got here
					message := "Error reading feed, unsubscribing for sanity: " + feed.URL + "\n"
					message = message + err.Error()
					s.ChannelMessageSend(feed.ChannelID, message )
					skip = true
				}

				for _, post := range feed.Posts {
					if post == item.Link {
						skip = true
					}
				}

				feed.LastItem = item.Link
				err = rss.UpdatePosts(feed)
				if err != nil{
					fmt.Println(err.Error())
					skip = true
				}

				if !skip {
					formatted := h.FormatRSSItem(feed.URL, item)
					s.ChannelMessageSend(feed.ChannelID, formatted)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}


func (h *RSSHandler) FormatRSSItem(url string, rssitem RSSItem) (formatted string) {

	if rssitem.Twitter {

		formatted = "Latest Tweet from " + rssitem.Author + "\n\n"
		formatted = formatted + rssitem.Link + "\n"

	} else if rssitem.Reddit {

		formatted = "New Subreddit Post from " + "https://www.reddit.com/user"+strings.TrimPrefix(rssitem.Author,"/u")+ " \n\n"
		formatted = formatted + rssitem.Title + "\n\n"
		//formatted = formatted + item.Content + "\n"
		//formatted = formatted + item.Description + "\n"
		formatted = formatted + rssitem.Link + "\n"
		formatted = formatted + rssitem.Published + "\n"

	} else if rssitem.Youtube {
		formatted = "Latest update from " + rssitem.Author + "\n"
		formatted = formatted + rssitem.Title + "\n"
		formatted = formatted + rssitem.Description + "\n"
		formatted = formatted + rssitem.Link + "\n"
	} else {

		formatted = "Latest update from " + url + "\n"
		formatted = formatted + rssitem.Author + "\n"
		formatted = formatted + rssitem.Title + "\n"
		formatted = formatted + rssitem.Content + "\n"
		//formatted = formatted + item.Description + "\n"
		formatted = formatted + rssitem.Link + "\n"
		formatted = formatted + rssitem.Published + "\n"

	}
	return formatted
}

// This will send, beware of that.
func (h *RSSHandler) GetAllLatest(s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	rss := RSS{db: h.db}
	feeds, err := rss.GetChannel(m.ChannelID)
	if err != nil {
		return err
	}


	for _, i := range feeds {

		item, err := rss.GetLatestItem(i.URL, m.ChannelID)
		if err != nil {
			return err
		}

		formatted := h.FormatRSSItem(i.URL, item)
		s.ChannelMessageSend(m.ChannelID, formatted)
		time.Sleep( 3 * time.Second)
	}

	return nil
}

func (h *RSSHandler) GetLatestItem(channel string, url string) (formatted string, err error) {
	rss := RSS{db: h.db}

	item, err := rss.GetLatestItem(url, channel)
	if err != nil {
		return formatted, err
	}

	formatted = h.FormatRSSItem(url, item)

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


func (h *RSSHandler) AddRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

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

	h.callback.Watch( h.ConfirmAddRSS, GetUUID(), m.Content, s, m)

}



func (h *RSSHandler) ConfirmAddRSS(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

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



func (h *RSSHandler) RemoveRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

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

	s.ChannelMessageSend(m.ChannelID, "Removing: " + m.Content + " Confirm? (Y/N)")

	h.callback.Watch( h.ConfirmRemoveRSS, GetUUID(), m.Content, s, m)

}



func (h *RSSHandler) ConfirmRemoveRSS(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp){
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	if m.Content == "Y" || m.Content == "y" {

		rss := RSS{db: h.db}
		err := rss.Unsubscribe(url, m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error Removing URL: " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: " + url )
		return
	}

	s.ChannelMessageSend(m.ChannelID, "RSS Remove Cancelled")
}