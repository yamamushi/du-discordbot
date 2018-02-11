package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

// RSSHandler struct
type RSSHandler struct {
	db         *DBHandler
	conf       *Config
	callback   *CallbackHandler
	dg         *discordgo.Session
	registry   *CommandRegistry
	foruminteg *ForumIntegration
}

// Init function
func (h *RSSHandler) Init() {
	h.RegisterCommands()
	h.foruminteg = &ForumIntegration{}

}

// RegisterCommands function
func (h *RSSHandler) RegisterCommands() (err error) {

	h.registry.Register("rss", "Manage the current channel's RSS Subscriptions", "rss add|remove|list")

	return nil

}

// Read function
func (h *RSSHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"rss") {
		if h.registry.CheckPermission("rss", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Moderator {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// ParseCommand function
func (h *RSSHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(command) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'rss' command")
		return
	}

	if command[1] == "add" && len(command) == 2 {
		s.ChannelMessageSend(m.ChannelID, "Please supply a feed URL: ")
		message := ""
		uuid, err := GetUUID()
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
			return
		}
		h.callback.Watch(h.AddRSS, uuid, message, s, m)
		return
	}

	if command[1] == "add" && len(command) > 2 {
		s.ChannelMessageSend(m.ChannelID, "Add RSS Feed: "+command[2]+" Confirm? (Y/N)")
		message := command[2]
		uuid, err := GetUUID()
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
			return
		}
		h.callback.Watch(h.ConfirmAddRSS, uuid, message, s, m)
		return
	}

	if command[1] == "remove" && len(command) == 2 {
		s.ChannelMessageSend(m.ChannelID, "Please supply a feed URL: ")
		message := ""
		uuid, err := GetUUID()
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
			return
		}
		h.callback.Watch(h.RemoveRSS, uuid, message, s, m)
		return
	}

	if command[1] == "remove" && len(command) > 2 {
		s.ChannelMessageSend(m.ChannelID, "Remove RSS Feed: "+command[2]+" Confirm? (Y/N)")
		message := command[2]
		uuid, err := GetUUID()
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
			return
		}
		h.callback.Watch(h.ConfirmRemoveRSS, uuid, message, s, m)
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

// UpdateRSSFeeds function
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

			if expires > h.conf.DUBotConfig.RSSTimeout*time.Minute {

				skip := false
				item, err := rss.GetLatestItem(feed.URL, feed.ChannelID)
				if err != nil {
					fmt.Println(err.Error())
					//rss.Unsubscribe(feed.URL, feed.ChannelID) // Unsubscribe us if we got here
					//message := "Error reading feed, unsubscribing for sanity: " + feed.URL + "\n"
					//message = message + err.Error()
					//s.ChannelMessageSend(feed.ChannelID, message)
					skip = true
				}

				for _, post := range feed.Posts {
					if post == item.Link {
						if !feed.RepeatPosts {
							skip = true
							item.Update = false
						} else {
							item.Update = true
						}
					}
				}

				if feed.LastItem == item.Link {
					skip = true
				}

				feed.LastItem = item.Link
				err = rss.UpdatePosts(feed)
				if err != nil {
					fmt.Println("RSS Update Error: " + err.Error())
					skip = true

				}

				if !skip {
					formatted := h.FormatRSSItem(feed.URL, item, feed.Title)
					s.ChannelMessageSend(feed.ChannelID, formatted)
				}
			}

			time.Sleep(10 * time.Second)
		}
	}
}

// FormatRSSItem function
func (h *RSSHandler) FormatRSSItem(url string, rssitem RSSItem, feedtitle string) (formatted string) {

	if rssitem.Twitter {

		formatted = "Latest Tweet from " + rssitem.Author + "\n\n"
		formatted = formatted + rssitem.Link + "\n"

	} else if rssitem.Reddit {

		formatted = ":postbox: New Subreddit Post from " + Bold("https://www.reddit.com/user"+strings.TrimPrefix(rssitem.Author, "/u")) + " \n\n"
		formatted = formatted + rssitem.Title + "\n\n"
		//formatted = formatted + item.Content + "\n"
		//formatted = formatted + item.Description + "\n"
		formatted = formatted + rssitem.Link + "\n"
		formatted = formatted + rssitem.Published + "\n"
		formatted = formatted + "============================\n"
	} else if rssitem.Youtube {
		//formatted = "Latest update from " + rssitem.Author + "\n"
		formatted = ":video_camera: New YouTube upload from " + rssitem.Author + "!\n"
		formatted = formatted + rssitem.Title + "\n"
		formatted = formatted + rssitem.Description + "\n"
		formatted = formatted + rssitem.Link + "\n"
		formatted = formatted + "============================\n"
	} else if rssitem.Forum {
		//formatted = "Latest update from " + rssitem.Author + "\n"
		feedtitle = strings.TrimSuffix(feedtitle, " Latest Topics")
		formatted = ":postbox: New unread post in " + Bold(feedtitle) + "\n"
		formatted = formatted + ":newspaper: || " + UnderlineBold(rssitem.Title) + " || \n" //+"<"+rssitem.Link+">\n\n"
		username, comment, commenturl, err := h.foruminteg.GetLatestCommentForThread(rssitem.Link)
		//fmt.Println("RSS debug: " + username + " " + comment + " " + commenturl)
		if err == nil {
			formatted = formatted + "New Comment from " + Bold(username) + ":\n"
			formatted = formatted + "```" + comment + "```\n"
			formatted = formatted + "Continue reading @ <" + commenturl + ">\n"
			formatted = formatted + "============================\n"
		}
	} else {
		formatted = ":postbox: New update: \n" // from " + url + "\n"
		formatted = formatted + rssitem.Author + "\n"
		formatted = formatted + rssitem.Title + "\n"
		formatted = formatted + rssitem.Content + "\n"
		//formatted = formatted + item.Description + "\n"
		formatted = formatted + rssitem.Link + "\n"
		formatted = formatted + rssitem.Published + "\n"
		formatted = formatted + "============================\n"
	}
	return formatted
}

// GetAllLatest function
// This will send, beware of that.
func (h *RSSHandler) GetAllLatest(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

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

		formatted := h.FormatRSSItem(i.URL, item, i.Title)
		s.ChannelMessageSend(m.ChannelID, formatted)
		time.Sleep(3 * time.Second)
	}
	return nil
}

// GetLatestItem function
func (h *RSSHandler) GetLatestItem(channel string, url string) (formatted string, err error) {
	rss := RSS{db: h.db}

	item, err := rss.GetLatestItem(url, channel)
	if err != nil {
		return formatted, err
	}

	formatted = h.FormatRSSItem(url, item, item.Title)

	return formatted, nil
}

// GetRSSList function
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

// AddRSS function
func (h *RSSHandler) AddRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// In this handler we don't do anything with the command string, instead we grab the response from m.Content

	// We do this to avoid having duplicate commands overrunning each other
	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	// A poor way of checking the validity of the RSS url for now
	if m.Content == "" {
		s.ChannelMessageSend(m.ChannelID, "Invalid Command Received")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Adding: "+m.Content+" Confirm? (Y/N)")


	uuid, err := GetUUID()
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
		return
	}
	h.callback.Watch(h.ConfirmAddRSS, uuid, m.Content, s, m)

}

// ConfirmAddRSS function
func (h *RSSHandler) ConfirmAddRSS(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	content := strings.Fields(m.Content)
	if len(content) > 0 {
		if content[0] == "Y" || content[0] == "y" {

			repeat := false
			if len(content) > 1 {
				if content[1] == "repeat" || content[1] == "Repeat" {
					repeat = true
				}
			}

			rss := RSS{db: h.db}
			err := rss.Subscribe(url, "", m.ChannelID, repeat)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error Subscribing to URL: "+err.Error())
				return
			}
			if repeat {
				s.ChannelMessageSend(m.ChannelID, "Repeating Selection Confirmed: "+url)
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: "+url)
			return
		}
	}

	s.ChannelMessageSend(m.ChannelID, "RSS Add Cancelled")
}

// RemoveRSS function
func (h *RSSHandler) RemoveRSS(command string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// In this handler we don't do anything with the command string, instead we grab the response from m.Content

	// We do this to avoid having duplicate commands overrunning each other
	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	// A poor way of checking the validity of the RSS url for now
	if m.Content == "" {
		s.ChannelMessageSend(m.ChannelID, "Invalid Command Received")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Removing: "+m.Content+" Confirm? (Y/N)")
	uuid, err := GetUUID()
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
		return
	}
	h.callback.Watch(h.ConfirmRemoveRSS, uuid, m.Content, s, m)

}

// ConfirmRemoveRSS function
func (h *RSSHandler) ConfirmRemoveRSS(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "RSS Command Cancelled")
		return
	}

	if m.Content == "Y" || m.Content == "y" {

		rss := RSS{db: h.db}
		err := rss.Unsubscribe(url, m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error Removing URL: "+err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: "+url)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "RSS Remove Cancelled")
}
