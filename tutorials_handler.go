package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"strconv"
)

// DevDiaryHandler struct
type TutorialHandler struct {
	user     *UserHandler
	conf     *Config
	db       *DBHandler
	registry *CommandRegistry
}

// Read function
func (h *TutorialHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !SafeInput(s, m, h.conf) {
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)

	command = strings.ToLower(command)

	for i := range payload {
		payload[i] = strings.ToLower(payload[i])
	}

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

	if command == "tutorial" {

		if !h.registry.CheckPermission("tutorial", m.ChannelID, user) {
			return
		}

		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command usage: tutorial <list> || tutorial <id>")
			return
		}

		if payload[0] == "list" {
			list := h.GetList()
			s.ChannelMessageSend(m.ChannelID, list)
			return
		}

		id, err := strconv.Atoi(payload[0])
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Invalid input supplied")
			return
		}

		url, title, date := h.GetForID(id)
		s.ChannelMessageSend(m.ChannelID, ":school: "+title+" || published "+date+" :\n"+url)
		return

	}

}


func (h *TutorialHandler) GetList() (list string) {

	message := "Dual Universe Interviews:```\n"
	message = message + "Title                                       | ID |\n"
	message = message + "--------------------------------------------------\n"
	message = message + "Tool & UI Basics                            | 1  |\n"
	message = message + "Atmospheric Ship Building                   | 2  |\n"
	message = message + "Interactive Elements & Linking              | 3  |\n"
	message = message + "Rights Management, Outposts & Territories   | 4  |\n"
	message = message + "Lua Scripting                               | 5  |\n"
	message = message + "Blueprint 101 (Preview)                     | 6  |\n"
	message = message + "--------------------------------------------------\n"
	message = message + "\n```"
	return message
}


func (h *TutorialHandler) GetForID(id int) (url string, title string, date string) {

	switch id {

	case 1:
		return "https://www.youtube.com/watch?v=wCpzLs4vlis",
		"Dual Universe Pre-Alpha Tutorial: Tool & UI Basics", "Oct 19 2017"
	case 2:
		return "https://www.youtube.com/watch?v=V3puZXotLIw",
		"Dual Universe Pre-Alpha Tutorial: Atmospheric Ship Building", "Oct 20 2017"
	case 3:
		return "https://www.youtube.com/watch?v=jPRx6WQlVQc",
		"Dual Universe Pre-Alpha Tutorial: Interactive Elements & Linking", "Oct 20 2017"
	case 4:
		return "https://www.youtube.com/watch?v=rdJQjiQXO8w",
		"Dual Universe Pre-Alpha Tutorial: Rights Management, Outposts & Territories", "Oct 20 2017"
	case 5:
		return "https://www.youtube.com/watch?v=sbvJPuo9npE",
		"Dual Universe Pre-Alpha Tutorial: Lua Scripting", "Nov 24 2017"
	case 6:
		return "https://www.youtube.com/watch?v=mEh3TzRPCyA",
			"Dual Universe Pre-Alpha Tutorial: Blueprint 101 (Preview)", "Feb 09 2017"
	default:
		return "No Record Found","nil","nil"
	}

}