package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

// DevDiaryHandler struct
type InterviewHandler struct {
	user     *UserHandler
	conf     *Config
	db       *DBHandler
	registry *CommandRegistry
}

// Read function
func (h *InterviewHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

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

	if command == "interview" {

		if !h.registry.CheckPermission("interview", m.ChannelID, user) {
			return
		}

		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command usage: interview <list> || interview <id>")
			return
		}

		if payload[0] == "list" {
			list := h.GetList()
			s.ChannelMessageSend(m.ChannelID, list)
			return
		}

		id, err := strconv.Atoi(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid input supplied")
			return
		}

		url, author, date := h.GetForID(id)
		s.ChannelMessageSend(m.ChannelID, "Interview with "+author+" uploaded on "+date+" :\n"+url)
		return

	}

}

func (h *InterviewHandler) GetList() (list string) {

	message := "Dual Universe Interviews:```\n"
	message = message + "Author            | ID | URL\n"
	message = message + "PCGamer           | 1  | https://www.youtube.com/watch?v=HKlx1ydHijA\n"
	message = message + "CaptainShack      | 2  | https://www.youtube.com/watch?v=m1WMwIDWFKI\n"
	message = message + "GrayStillPlays    | 3  | https://www.youtube.com/watch?v=BFM4sTFMU1U\n"
	message = message + "DUExplorers       | 4  | https://www.youtube.com/watch?v=mQAeDxhbiyo\n"
	message = message + "SandboxerTV       | 5  | https://www.youtube.com/watch?v=WliVfDzWEaA\n"
	message = message + "Cobra TV          | 6  | https://www.youtube.com/watch?v=DD0gdy7ABQ8\n"
	message = message + "GamerZakh         | 7  | https://www.youtube.com/watch?v=SlJyYzvGHu8\n"
	message = message + "Gaiscioch         | 8  | https://www.youtube.com/watch?v=MTSyWYeh3Sk\n"
	message = message + "GameSpot          | 9  | https://www.youtube.com/watch?v=uE5-CxQ8wzI\n"
	message = message + "Captain Jack      | 10 | https://www.youtube.com/watch?v=TFUKzex-vkc\n"
	message = message + "markeedragon      | 11 | https://www.youtube.com/watch?v=C5WYWzSfwOk\n"
	message = message + "DUExplorers       | 12 | https://www.youtube.com/watch?v=H9Y0YmqGYDM\n"
	message = message + "Space Game Junkie | 13 | https://www.youtube.com/watch?v=m8Kw1Af6LKk\n"
	message = message + "DM21 Gaming       | 14 | https://www.youtube.com/watch?v=VVkbdfxiKxM\n"
	message = message + "IGN               | 15 | https://www.youtube.com/watch?v=W8cLaCn5A94\n"
	message = message + "\n```"
	return message
}

func (h *InterviewHandler) GetForID(id int) (url string, author string, date string) {

	switch id {

	case 1:
		return "https://www.youtube.com/watch?v=HKlx1ydHijA", "PCGamer", "Jun 16 2016"
	case 2:
		return "https://www.youtube.com/watch?v=m1WMwIDWFKI", "CaptainShack", "Jun 22 2016"
	case 3:
		return "https://www.youtube.com/watch?v=BFM4sTFMU1U", "GrayStillPlays", "Aug 06 2016"
	case 4:
		return "https://www.youtube.com/watch?v=mQAeDxhbiyo", "DUExplorers", "Sep 05 2016"
	case 5:
		return "https://www.youtube.com/watch?v=WliVfDzWEaA", "SandboxerTV", "Sep 06 2016 "
	case 6:
		return "https://www.youtube.com/watch?v=DD0gdy7ABQ8", "Cobra TV", "Sep 07 2016"
	case 7:
		return "https://www.youtube.com/watch?v=SlJyYzvGHu8", "GamerZakh", "Sep 07 2016"
	case 8:
		return "https://www.youtube.com/watch?v=MTSyWYeh3Sk", "Gaiscioch", "Sep 07 2016"
	case 9:
		return "https://www.youtube.com/watch?v=uE5-CxQ8wzI", "GameSpot", "Sep 07 2016"
	case 10:
		return "https://www.youtube.com/watch?v=TFUKzex-vkc", "Captain Jack", "Sep 08 2016"
	case 11:
		return "https://www.youtube.com/watch?v=C5WYWzSfwOk", "markeedragon", "Sep 17 2016"
	case 12:
		return "https://www.youtube.com/watch?v=H9Y0YmqGYDM", "DUExplorers", "Sep 17 2016"
	case 13:
		return "https://www.youtube.com/watch?v=m8Kw1Af6LKk", "Space Game Junkie", "Sep 27 2016"
	case 14:
		return "https://www.youtube.com/watch?v=VVkbdfxiKxM", "DM21 Gaming", "Oct 08 2016"
	case 15:
		return "https://www.youtube.com/watch?v=W8cLaCn5A94", "IGN", "Mar 04 2017"
	default:
		return "No Record Found", "nil", "nil"
	}

}
