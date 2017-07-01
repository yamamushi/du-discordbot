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
	command *CommandHandler
	registry *CommandRegistry
	logger *Logger
}

func (h *MainHandler) Init() error {
	// DO NOT add anything above this line!!
	fmt.Println("Initializing Main Handler")
	// Add our main handler -
	h.dg.AddHandler(h.Read)
	h.registry = h.command.registry


	// Add new handlers below this line //
	// Create our RSS handler
	fmt.Println("Adding RSS Handler")
	rss := RSSHandler{db: h.db, conf: h.conf, callback: h.callback, dg: h.dg, registry: h.registry}
	rss.Init()
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

	h.RegisterCommands()

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
		//fmt.Println("Error finding user")
		return
	}

	message := strings.Fields(m.Content)
	if len(message) < 1 {
		fmt.Println(m.Content)
		return
	}

	command := message[0]

	// If the message is "ping" reply with "Pong!"
	if command == cp + "ping" {
		if CheckPermissions("ping", m.ChannelID, &user, s, h.command){
			s.ChannelMessageSend(m.ChannelID, "Pong!")
			return
		}
	}

	// If the message is "pong" reply with "Ping!"
	if command == cp + "pong" {
		if CheckPermissions("pong", m.ChannelID, &user, s, h.command){
			s.ChannelMessageSend(m.ChannelID, "Ping!")
			return
		}
	}

	if command == cp + "help" {
		s.ChannelMessageSend(m.ChannelID, "https://imgfave.azureedge.net/image_cache/140248453569251_animate.gif")
	}

	if command == cp + "follow" {
		if CheckPermissions("follow", m.ChannelID, &user, s, h.command){
			s.ChannelMessageSend(m.ChannelID, "Not yet implemented!")
			return
		}

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


func (h *MainHandler) RegisterCommands() (err error) {

	h.registry.Register("follow", "Follow a DU forum user. Updates will be sent via pm", "follow <forum name>")
	h.registry.Register("ping", "Ping command", "ping")
	h.registry.Register("pong", "Pong command", "pong")

	return nil

}
