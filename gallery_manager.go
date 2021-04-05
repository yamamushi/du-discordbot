package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

type GalleryManager struct {
	conf     *Config
	registry *CommandRegistry
	db       *DBHandler
	gallery  *GalleryDB
	userdb   *UserHandler
}

// Init function
func (h *GalleryManager) Init() {
	_ = h.RegisterCommands()
	h.gallery = &GalleryDB{db: h.db}
}

// RegisterCommands function
func (h *GalleryManager) RegisterCommands() (err error) {
	h.registry.Register("gallery", "Gallery Manager", "See help page ```~gallery help```")
	return nil
}

// Read function
func (h *GalleryManager) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"gallery") {
		if h.registry.CheckPermission("gallery", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Admin {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// ParseCommand function
func (h *GalleryManager) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	_, payload := SplitPayload(commandlist)
	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Error: Command expects an argument")
		return
	}
	for i := range payload {
		payload[i] = strings.ToLower(payload[i])
	}

	// Allow text from a user
	if payload[0] == "help" {
		help := "Gallery Manager ```" +
			"whitelist: <channel> <user>\n" +
			"blacklist: <channel> <user>\n" +
			"enable: <channel>\n" +
			"disable: <channel>```"
		s.ChannelMessageSend(m.ChannelID, help)
		return
	}

	// Allow text from a user
	if payload[0] == "whitelist" {
		if len(payload) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Error: Command expects a channel and a user mention argument")
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Error: Command expects a channel and a user mention argument")
			return
		}

		channelid := CleanChannel(payload[1])
		_, err := strconv.Atoi(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel selected")
			return
		}

		config, err := h.gallery.GetConfigFromDB(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: A record was not found for "+payload[1])
			return
		}

		config.Whitelist = AppendStringIfMissing(config.Whitelist, m.Mentions[0].ID)
		err = h.gallery.UpdateConfig(config)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Error saving config to DB: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" was whitelisted in "+payload[1])
		return
	}

	// Disable user text permissions (default is blacklisted)
	if payload[0] == "blacklist" {
		if len(payload) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Error: Command expects a channel and a user mention argument")
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Error: Command expects a channel and a user mention argument")
			return
		}

		channelid := CleanChannel(payload[1])
		_, err := strconv.Atoi(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel selected")
			return
		}

		config, err := h.gallery.GetConfigFromDB(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: A record was not found for "+payload[1])
			return
		}

		config.Whitelist = RemoveStringFromSlice(config.Whitelist, m.Mentions[0].ID)
		err = h.gallery.UpdateConfig(config)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Error saving config to DB: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" was removed from the whitelist in "+payload[1])
		return
	}

	// Enable the gallery manager
	if payload[0] == "enable" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Error: Command expects a channel argument")
			return
		}
		channelid := CleanChannel(payload[1])
		_, err := strconv.Atoi(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel selected")
			return
		}

		config, err := h.gallery.GetConfigFromDB(channelid)
		if err != nil {
			config = GalleryConfig{ChannelID: channelid, Enabled: true}
			err = h.gallery.AddConfigToDB(config)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Error saving config to DB: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Gallery enabled for "+payload[1])
			return
		}
		config.Enabled = true
		err = h.gallery.UpdateConfig(config)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Error saving config to DB: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Gallery enabled for "+payload[1])
		return
	}

	// Disable the gallery manager
	if payload[0] == "disable" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Error: Command expects a channel argument")
			return
		}
		channelid := CleanChannel(payload[1])
		_, err := strconv.Atoi(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel selected")
			return
		}

		config, err := h.gallery.GetConfigFromDB(channelid)
		if err != nil {
			config = GalleryConfig{ChannelID: channelid, Enabled: false}
			err = h.gallery.AddConfigToDB(config)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Error saving config to DB: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Gallery enabled for "+payload[1])
			return
		}
		config.Enabled = false
		err = h.gallery.UpdateConfig(config)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: Error saving config to DB: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Gallery disabled for "+payload[1])
		return
	}
}

func (h *GalleryManager) Watch(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	// Allow admins to pass commands
	if strings.HasPrefix(m.Content, h.conf.DUBotConfig.CP) {
		user, err := h.userdb.GetUser(m.Author.ID)
		if err == nil {
			if user.Admin {
				return
			}
		}
	}

	config, err := h.gallery.GetConfigFromDB(m.ChannelID)
	if err != nil {
		return
	}
	if !config.Enabled {
		return
	}

	allowedmessage := false
	if len(m.Attachments) > 0 {
		allowedmessage = true
	}
	if len(m.Embeds) > 0 {
		allowedmessage = true
	}

	content := strings.Split(m.Message.Content, " ")
	for _, word := range content {
		if IsValidUrl(word) {
			if IsReachableURL(word) {
				allowedmessage = true
				break
			}
		}
	}

	for _, userid := range config.Whitelist {
		if m.Author.ID == userid {
			allowedmessage = true
			break
		}
	}

	if !allowedmessage {
		discordmember, err := s.GuildMember(h.conf.DiscordConfig.GuildID, m.Author.ID)
		if err != nil {
			return
		}
		userroles := discordmember.Roles
		for _, role := range userroles {
			rolename, err := getRoleNameByID(role, h.conf.DiscordConfig.GuildID, s)
			if err != nil {
				return
			}

			if rolename == "NQ-Staff" || rolename == "Discord Staff" {
				allowedmessage = true
				break
			}
		}
	}

	if !allowedmessage {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		privatechannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" - Your message was removed automatically. Posting text-only messages is disabled in the gallery.")
			return
		}
		s.ChannelMessageSend(privatechannel.ID, "This is a notification that one of your messages was removed automatically. Text-only messages are disabled in the gallery.")
		return
	}

	return
}
