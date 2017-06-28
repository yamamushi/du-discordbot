package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

type MainHandler struct {

	db *DBHandler
	conf *Config
	dg *discordgo.Session
	callback *CallbackHandler
	perm *PermissionsHandler
	user *UserHandler
}

func (h *MainHandler) Init() error {
	// DO NOT add anything above this line!!
	fmt.Println("Initializing Main Handler")
	// Add our main handler -
	h.dg.AddHandler(h.Read)

	// Add new handlers below this line //
	// Create our RSS handler
	fmt.Println("Adding RSS Handler")
	rss := RSSHandler{db: h.db, conf: h.conf, callback: h.callback, dg: h.dg}
	h.dg.AddHandler(rss.Read)
	go rss.UpdateRSSFeeds(h.dg)

	// Open a websocket connection to Discord and begin listening.
	fmt.Println("Opening Connection to Discord")
	err := h.dg.Open()
	if err != nil {
		fmt.Println("Error Opening Connection: ", err)
		return err
	}
	fmt.Println("Connection Established")


	err = h.PostInit(h.dg)

	if err != nil {
		fmt.Println("Error during Post-Init")
		return err
	}

	fmt.Println("Main Handler Initialized")
	return nil
}


// Just some quick things to run after our websocket has been setup and opened

func (h *MainHandler) PostInit(dg *discordgo.Session) error {
	fmt.Println("Running Post-Init")

	// Update our default playing status
	fmt.Println("Updating Discord Status")
	err := h.dg.UpdateStatus(0, h.conf.DUBotConfig.Playing)
	if err != nil {
		fmt.Println("error updating now playing,", err)
		return err
	}

	fmt.Println("Post-Init Complete")
	return nil
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (h *MainHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
	// very important to set this first!
	cp := h.conf.DUBotConfig.CP

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println("Error finding user")
		return
	}

	message := strings.Fields(m.Content)

	command := message[0]

	// If the message is "ping" reply with "Pong!"
	if command == cp + "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if command == cp + "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if command == cp + "help" {
		s.ChannelMessageSend(m.ChannelID, "http://imgfave.com/collection/307305/Reaction-GIFs-no")
	}

	if command == cp + "follow" {

		if !user.Admin {
			return
		}
		if len(command) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command usage: follow <user>")
		}

		forum := ForumIntegration{}
		forum.FollowUser(message[1])
		s.ChannelMessageSend(m.ChannelID, "Callback launched")
	}

}

