package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// ComputerHandler struct
type ComputerHandler struct {
	user     *UserHandler
	conf     *Config
	db       *DBHandler
	registry *CommandRegistry
	callback *CallbackHandler
}

// Read function
func (h *ComputerHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !SafeInput(s, m, h.conf) {
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)

	h.user.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if !user.Citizen {
		return
	}
	/*
		command = payload[0]
		payload = RemoveStringFromSlice(payload, command)
	*/

	if command == "computer" || command == "Computer" {

		if !h.registry.CheckPermission("computer", m.ChannelID, user) {
			return
		}

		if len(payload) < 2 {
			return
		}
		if payload[0] == "nude" && payload[1] == "tayne" {
			s.ChannelMessageSend(m.ChannelID, "This is not suitable for work. Are you sure?")
			uuid, err := GetUUID()
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID")
				return
			}
			h.callback.Watch(h.TayneOhGod, uuid, "", s, m)
			return
		}

		if len(payload) < 3 {
			return
		}
		if payload[0] == "add" && payload[1] == "sequence:" && payload[2] == "OYSTER" {
			s.ChannelMessageSend(m.ChannelID, "http://i.imgur.com/LGnzAXN.gif")
			return
		}

		if payload[0] == "and" && payload[1] == "a" && payload[2] == "flarhgunnstow?" {
			s.ChannelMessageSend(m.ChannelID, "Yes. http://i.imgur.com/zlz25iD.gif")
			return
		}

		if len(payload) < 5 {
			return
		}
		if payload[0] == "load" && payload[1] == "up" && payload[2] == "celery" && payload[3] == "man" && payload[4] == "please" {
			s.ChannelMessageSend(m.ChannelID, "Yes "+m.Author.Mention()+" https://www.tenor.co/zSBS.gif")
			return
		}

		if len(payload) < 6 {
			return
		}
		if payload[0] == "could" && payload[1] == "you" && payload[2] == "kick" && payload[3] == "up" && payload[4] == "the" && payload[5] == "4d3d3d3?" {
			s.ChannelMessageSend(m.ChannelID, "4D3d3d3 Engaged "+m.Author.Mention()+" https://www.tenor.co/uk58.gif")
			return
		}

		if payload[0] == "could" && payload[1] == "I" && payload[2] == "see" && payload[3] == "a" && payload[4] == "hat" && payload[5] == "wobble?" {
			s.ChannelMessageSend(m.ChannelID, "Yes. http://i.imgur.com/QVnGKCH.gif")
			return
		}

		if payload[0] == "do" && payload[1] == "we" && payload[2] == "have" && payload[3] == "any" &&
			payload[4] == "new" && payload[5] == "sequences?" {
			s.ChannelMessageSend(m.ChannelID, "I have a BETA sequence\nI have been working on\nWould you like to see it?")

			uuid, err := GetUUID()
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: " + err.Error())
				return
			}
			h.callback.Watch(h.TayneResponse, uuid, "", s, m)
			return
		}

		if len(payload) < 8 {
			return
		}

		if payload[0] == "give" && payload[1] == "me" && payload[2] == "a" && payload[3] == "print" &&
			payload[4] == "out" && payload[5] == "of" && payload[6] == "oyster" && payload[7] == "smiling" {
			s.ChannelMessageSend(m.ChannelID, "okay. https://i.imgur.com/Qrhid0G.png")
			return
		}

		if len(payload) < 9 {
			return
		}
		if payload[0] == "is" && payload[1] == "there" && payload[2] == "any" && payload[3] == "way" &&
			payload[4] == "to" && payload[5] == "generate" && payload[6] == "a" && payload[7] == "nude" && payload[8] == "tayne?" {
			s.ChannelMessageSend(m.ChannelID, "Not Computing. Please repeat.")
			return
		}
	}

}

// TayneResponse function
func (h *ComputerHandler) TayneResponse(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Computer Command Cancelled")
		return
	}

	content := strings.Fields(m.Content)
	if len(content) > 0 {
		if content[0] == "alright" || content[0] == "Alright" {
			s.ChannelMessageSend(m.ChannelID, "Okay http://i.imgur.com/5K4qcE4.gif")
			return
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Computer Input Cancelled")
}

// TayneOhGod function
func (h *ComputerHandler) TayneOhGod(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Computer Command Cancelled")
		return
	}

	content := strings.Fields(m.Content)
	if len(content) > 0 {
		if content[0] == "yes" || content[0] == "mhmm" || content[0] == "yep" {
			s.ChannelMessageSend(m.ChannelID, "Okay https://www.tenor.co/Fcsn.gif")
			return
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Computer Input Cancelled")
}
