package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
)

type ChannelHandler struct {
	db        *DBHandler
	conf      *Config
	registry  *CommandRegistry
	channeldb *ChannelDB
	user      *UserHandler
	logchan   chan string
}

func (h *ChannelHandler) Init() {

	h.channeldb = new(ChannelDB)
	h.channeldb.db = h.db

}

func (h *ChannelHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Check for safety
	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		return
	}
	if !user.CheckRole("admin") {
		return
	}

	command, message := CleanCommand(m.Content, h.conf)

	if command == "channel" {
		h.ReadCommand(message, s, m)
	}
}

func (h *ChannelHandler) ReadCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(message) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<channel> requires an argument")
		return
	}

	command := message[0]
	payload := RemoveStringFromSlice(message, command)

	if command == "info" {
		h.Info(payload, s, m)
		return
	}
	if command == "set" {
		h.Set(payload, s, m)
		return
	}
	if command == "unset" {
		h.Unset(payload, s, m)
		return
	}
	if command == "group" {
		h.ReadGroup(payload, s, m)
		return
	}
}

func (h *ChannelHandler) Info(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	var channelid string
	var formattedoutput string
	if len(payload) < 1 {
		channelid = m.ChannelID
	}
	if len(payload) > 0 {
		channelid = CleanChannel(payload[0])
		_, err := strconv.Atoi(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid channel selected")
		}
	}

	formattedchannel, err := MentionChannel(channelid, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	h.channeldb.CreateIfNotExists(channelid)
	channelrecord, err := h.channeldb.GetChannel(channelid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	var box string

	if channelrecord.HQ {
		box = box + "Is HQ\n"
	}

	if channelrecord.IsBotLog {
		box = box + "Is Bot Log\n"
	}
	if channelrecord.IsBankLog {
		box = box + "Is Bank Log\n"
	}
	if channelrecord.IsPermissionLog {
		box = box + "Is Permission Log\n"
	}
	if channelrecord.IsMusicRoom {
		box = box + "Is Music Room\n"
	}

	var formattedgroups string
	for i, group := range channelrecord.Groups {
		if i == len(channelrecord.Groups)-1 {
			formattedgroups = formattedgroups + group
		} else {
			formattedgroups = formattedgroups + group + ", "
		}
	}

	if formattedgroups != "" {
		box = box + "\nGroups: " + formattedgroups
	}

	if box != "" {
		formattedoutput = ":\n Info for " + formattedchannel + "\n"
		formattedoutput = formattedoutput + "```\n" + box + "\n```"
	} else {
		formattedoutput = "No information found for " + formattedchannel
	}

	s.ChannelMessageSend(m.ChannelID, formattedoutput)
}

func (h *ChannelHandler) Set(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	var channelid string
	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<set> requires an argument")
		return
	}
	if len(payload) == 1 {

		channelid = m.ChannelID
		dgchannel, err := s.Channel(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if payload[0] == "botlog" {
			err := h.channeldb.SetBotLog(channelid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "<#"+dgchannel.ID+"> has been set to the bot log")
			return
		}
		if payload[0] == "permissionlog" {
			err := h.channeldb.SetPermissionLog(channelid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "<#"+dgchannel.ID+"> has been set to the permission log")
			return
		}
		if payload[0] == "banklog" {
			err := h.channeldb.SetBankLog(channelid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "<#"+dgchannel.ID+"> has been set to the bank log")
			return
		}
		if payload[0] == "hq" {
			user, err := h.user.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			if !user.Owner {
				s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/eYcGQ5t.gif")
				return
			}

			err = h.channeldb.SetHQ(channelid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "HQ has been assigned to <#"+dgchannel.ID+">")
			return
		}
		if payload[0] == "musicroom" {
			user, err := h.user.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			if !user.Owner {
				s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/eYcGQ5t.gif")
				return
			}

			err = h.channeldb.SetMusicRoom(channelid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Music Room has been assigned to <#"+dgchannel.ID+">")
			return
		}
		s.ChannelMessageSend(m.ChannelID, payload[0]+" is not a valid room type")
		return
	}
	if len(payload) > 1 {
		channelmention := payload[1]
		channelid = CleanChannel(channelmention)
		_, err := strconv.Atoi(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Ivnvalid channel selected")
		}

		dgchannel, err := s.Channel(channelid)
		if err != nil {

			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		if payload[0] == "botlog" {
			err := h.channeldb.SetBotLog(dgchannel.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "<#"+dgchannel.ID+"> has been set to the bot log")
			return
		}
		if payload[0] == "permissionlog" {
			err := h.channeldb.SetPermissionLog(dgchannel.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "<#"+dgchannel.ID+"> has been set to the permission log")
			return
		}
		if payload[0] == "banklog" {
			err := h.channeldb.SetBankLog(dgchannel.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "<#"+dgchannel.ID+"> has been set to the bank log")
			return
		}
		if payload[0] == "hq" {
			user, err := h.user.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			if !user.Owner {
				s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/eYcGQ5t.gif")
				return
			}

			err = h.channeldb.SetHQ(dgchannel.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "HQ has been assigned to <#"+dgchannel.ID+">")
			return
		}
		if payload[0] == "musicroom" {
			user, err := h.user.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			if !user.Owner {
				s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/eYcGQ5t.gif")
				return
			}

			err = h.channeldb.SetMusicRoom(dgchannel.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Music Room has been assigned to <#"+dgchannel.ID+">")
			return
		}
		s.ChannelMessageSend(m.ChannelID, payload[0]+" is not a valid room type")
	}
}

func (h *ChannelHandler) Unset(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<unset> requires an argument")
		return
	}

	if payload[0] == "botlog" {
		err := h.channeldb.RemoveBotLog()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Botlog disabled")
		return
	}
	if payload[0] == "permissionlog" {
		err := h.channeldb.RemovePermissionLog()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Permission log disabled")
		return
	}
	if payload[0] == "banklog" {
		err := h.channeldb.RemoveBankLog()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Bank log disabled")
		return
	}
	if payload[0] == "hq" {
		user, err := h.user.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/eYcGQ5t.gif")
			return
		}

		err = h.channeldb.RemoveHQ()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "HQ has been unassigned")
		return
	}
	if payload[0] == "musicroom" {
		user, err := h.user.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/eYcGQ5t.gif")
			return
		}

		err = h.channeldb.RemoveMusicRoom()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Music Room has been unassigned")
		return
	}
	s.ChannelMessageSend(m.ChannelID, payload[0]+" is not a valid room type")
	return

}

func (h *ChannelHandler) GetBotLogChannel() (channelid string, err error) {
	return h.channeldb.GetBotLog()
}

func (h *ChannelHandler) GetPermissionLogChannel() (channelid string, err error) {
	return h.channeldb.GetPermissionLog()
}

func (h *ChannelHandler) GetBankLogChannel() (channelid string, err error) {
	return h.channeldb.GetBankLog()
}

func (h *ChannelHandler) GetHQChannel() (channelid string, err error) {
	return h.channeldb.GetHQ()
}

func (h *ChannelHandler) GetMusicRoomChannel() (channelid string, err error) {
	return h.channeldb.GetMusicRoom()
}

func (h *ChannelHandler) ReadGroup(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<group> requires an argument")
		return
	}

	command, message := SplitPayload(payload)

	if len(message) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<"+command+"> requires an argument")
		return
	}
	if command == "add" {
		h.AddGroup(message, s, m)
		return
	}
	if command == "remove" {
		h.RemoveGroup(message, s, m)
		return
	}
}

func (h *ChannelHandler) RemoveGroup(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(payload) < 2 {
		s.ChannelMessageSend(m.ChannelID, "<remove> requires two arguments")
		return
	}

	channelid := CleanChannel(payload[1])
	_, err := strconv.Atoi(channelid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid channel selected")
		return
	}
	formattedchannel, err := MentionChannel(channelid, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid channel selected")
		return
	}

	err = h.channeldb.RemoveGroup(channelid, payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, formattedchannel+" was removed from the "+payload[0]+" group.")
}

func (h *ChannelHandler) AddGroup(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(payload) < 2 {
		s.ChannelMessageSend(m.ChannelID, "<add> requires two arguments")
		return
	}

	channelid := CleanChannel(payload[1])
	_, err := strconv.Atoi(channelid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid channel selected")
		return
	}
	formattedchannel, err := MentionChannel(channelid, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid channel selected")
		return
	}

	err = h.channeldb.AddGroup(channelid, payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, formattedchannel+" was added to the "+payload[0]+" group.")

}

func (h *ChannelHandler) CheckPermission(channelid string, user *User) bool {

	record, err := h.channeldb.GetChannel(channelid)
	if err != nil {
		return false
	}

	for _, channelgroup := range record.Groups {
		if user.CheckRole(channelgroup) {
			return true
		}
	}

	return false
}
