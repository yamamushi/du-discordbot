package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

// PrimaryHandler struct
type PrimaryHandler struct {
	db          *DBHandler
	conf        *Config
	dg          *discordgo.Session
	callback    *CallbackHandler
	perm        *PermissionsHandler
	user        *UserHandler
	command     *CommandHandler
	registry    *CommandRegistry
	logchan     chan string
	bankhandler *BankHandler
	channel     *ChannelHandler
}

// Init function
func (h *PrimaryHandler) Init() error {
	// DO NOT add anything above this line!!
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

	fmt.Println("Adding Chess Handler")
	chess := ChessHandler{db: h.db, conf: h.conf, logchan: h.logchan, wallet: h.bankhandler.wallet,
		bank: h.bankhandler, command: h.command.registry, user: h.user}
	chess.Init()
	h.dg.AddHandler(chess.Read)

	fmt.Println("Adding Utilities Handler")
	utilities := UtilitiesHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry, logchan: h.logchan, callback: h.callback}
	h.dg.AddHandler(utilities.Read)

	fmt.Println("Adding Computer command Handler")
	computerhandler := ComputerHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry, callback: h.callback}
	h.dg.AddHandler(computerhandler.Read)

	fmt.Println("Adding Lua Handler")
	luahandler := LuaHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry}
	h.dg.AddHandler(luahandler.Read)

	fmt.Println("Adding DevDiary Handler")
	devdiary := DevDiaryHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry}
	h.dg.AddHandler(devdiary.Read)

	fmt.Println("Adding Interview Handler")
	interviews := InterviewHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry}
	h.dg.AddHandler(interviews.Read)

	fmt.Println("Adding Tutorial Handler")
	tutorials := TutorialHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry}
	h.dg.AddHandler(tutorials.Read)

	fmt.Println("Adding Meme Handler")
	memes := MemeHandler{}
	h.dg.AddHandler(memes.Read)

	fmt.Println("Adding Backer Handler")
	backer := BackerHandler{db: h.db, callback: h.callback, conf: h.conf}
	backer.Init()
	h.dg.AddHandler(backer.Read)

	fmt.Println("Adding Notifications Handler")
	notifications := NotificationsHandler{db: h.db, callback: h.callback, conf: h.conf, registry: h.command.registry}
	notifications.Init()
	h.dg.AddHandler(notifications.Read)
	go notifications.CheckNotifications(h.dg)

	//fmt.Println("Adding Music Handler")
	//musichandler := MusicHandler{db: h.db, user: h.user, registry: h.command.registry,
	//	wallet: h.bankhandler.wallet, channel: h.channel, conf: h.conf}
	//musichandler.Init()
	//h.dg.AddHandler(musichandler.Read)

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

	return nil
}

// PostInit function
// Just some quick things to run after our websocket has been setup and opened
func (h *PrimaryHandler) PostInit(dg *discordgo.Session) error {
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

// Read function
func (h *PrimaryHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	if err != nil {
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
	if command == cp+"ping" {
		if CheckPermissions("ping", m.ChannelID, &user, s, h.command) {
			s.ChannelMessageSend(m.ChannelID, "Pong!")
			return
		}
	}

	// If the message is "pong" reply with "Ping!"
	if command == cp+"pong" {
		if CheckPermissions("pong", m.ChannelID, &user, s, h.command) {
			s.ChannelMessageSend(m.ChannelID, "Ping!")
			return
		}
	}

	if command == cp+"time" {
		s.ChannelMessageSend(m.ChannelID, "Current UTC Time: "+time.Now().UTC().Format("2006-01-02 15:04:05"))
		return
	}

	if command == cp+"help" {
		s.ChannelMessageSend(m.ChannelID, "https://github.com/yamamushi/du-discordbot#table-of-contents")
		return
	}

	/*
		if command == cp+"follow" {
			if CheckPermissions("follow", m.ChannelID, &user, s, h.command) {
				s.ChannelMessageSend(m.ChannelID, "Not yet implemented!")
				return
			}

			if !user.Admin {
				return
			}
			if len(command) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Command usage: follow <user>")
				return
			}

			forum := ForumIntegration{}
			forum.FollowUser(message[1])
			s.ChannelMessageSend(m.ChannelID, "Callback launched")
			return
		}*/
}

// RegisterCommands function
func (h *PrimaryHandler) RegisterCommands() (err error) {

	h.registry.Register("follow", "Follow a DU forum user. Updates will be sent via pm", "follow <forum name>")
	h.registry.Register("ping", "Ping command", "ping")
	h.registry.Register("pong", "Pong command", "pong")
	h.registry.Register("transfer", "Transfer credits to another user", "transfer 100 @<user>")
	h.registry.Register("balance", "Display user balance", "balance")
	h.registry.Register("addbalance", "-?-", "-?-")
	h.registry.Register("chess", "du-discordbot chess", "chess")
	h.registry.Register("computer", "celery man please", "computer")
	h.registry.Register("devdiary", "Display a dev diary for a given month", "devdiary <month>")
	h.registry.Register("interview", "Display an interview for a given id", "interview <id> || <list>")
	h.registry.Register("tutorial", "Display a tutorial for a given id", "tutorial <id> || <list>")

	return nil
}
