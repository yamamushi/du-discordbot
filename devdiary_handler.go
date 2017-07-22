package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// DevDiaryHandler struct
type DevDiaryHandler struct {
	user     *UserHandler
	conf     *Config
	db       *DBHandler
	registry *CommandRegistry
}

// Read function
func (h *DevDiaryHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

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

	if command == "devdiary" {

		if !h.registry.CheckPermission("devdiary", m.ChannelID, user) {
			return
		}

		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command usage: devdiary <month> <year>")
			return
		}

		month := payload[0]
		year := payload[1]

		if year == "2016" || year == "16" {
			if month == "december" || month == "12" || month == "dec" {
				diarylink := "DevDiary for December 2016:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=WTvT4BAg7RI"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			s.ChannelMessageSend(m.ChannelID, "No DevDiary for "+month+" "+year+" found.")
			return
		}

		if year == "2017" || year == "17" {
			if month == "january" || month == "1" || month == "jan" {
				diarylink := "DevDiary for January 2017:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=iUTyiMjjf7w"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			if month == "february" || month == "2" || month == "feb" {
				diarylink := "DevDiary for February 2017:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=rNEgXg9vTik"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			if month == "march" || month == "3" || month == "mar" {
				diarylink := "DevDiary for March 2017:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=5yG5DtZcdd8"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			if month == "april" || month == "4" || month == "apr" {
				diarylink := "DevDiary for April 2017:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=GVJyfKNeOao"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			if month == "may" || month == "5" {
				diarylink := "DevDiary for May 2017:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=Vce1qcnyTpE"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			if month == "june" || month == "6" || month == "jun" {
				diarylink := "DevDiary for June 2017:\n"
				diarylink = diarylink + "https://www.youtube.com/watch?v=pz-XcdbaC3E"
				s.ChannelMessageSend(m.ChannelID, diarylink)
				return
			}
			s.ChannelMessageSend(m.ChannelID, "No DevDiary for "+month+" "+year+" found.")
			return
		}

	}

}
