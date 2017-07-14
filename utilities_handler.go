package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lunixbochs/vtclean"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UtilitiesHandler struct
type UtilitiesHandler struct {
	user     *UserHandler
	db       *DBHandler
	conf     *Config
	registry *CommandRegistry
	logchan  chan string
	callback *CallbackHandler
}

// ShortURLResponse struct
type ShortURLResponse struct {
	Short map[string]string `json:"/short"`
}

// Read function
func (h *UtilitiesHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

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

	if command == "unfold" || command == "unshorten" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<unfold> expects an argument")
			return
		}
		response, err := h.UnfoldURL(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not unfold "+payload[0]+" : "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, payload[0]+" unfolds to ```\n"+response+"\n```")
		return
	}
	if command == "moon" {
		message, err := h.GetMoon()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not get moon: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Current Moon From Earth:\n"+message)
		return

	}
	if command == "countdown" {
		s.ChannelMessageSend(m.ChannelID, h.GetCountdownStatus())
		return
	}
	if command == "pledging" || command == "pledges" || command == "crowdfunding" || command == "founderspacks" {

		s.ChannelMessageSend(m.ChannelID, h.GetPledgingStatus())
		return
	}
	if command == "computer" || command == "Computer" {

		if len(payload) < 2 {
			return
		}
		if payload[0] == "nude" && payload[1] == "tayne" {
			s.ChannelMessageSend(m.ChannelID, "This is not suitable for work. Are you sure?")
			h.callback.Watch(h.TayneOhGod, GetUUID(), "", s, m)
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
			h.callback.Watch(h.TayneResponse, GetUUID(), "", s, m)
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

// UnfoldURL function
func (h *UtilitiesHandler) UnfoldURL(input string) (output string, err error) {

	// The first step is to use golang’s http module to get the response:
	res, err := http.Get("http://x.datasig.io/short?url=" + input)
	if err != nil {
		return output, err
	}

	// Assuming you didnt see a panic call, the response to this http call is
	// being stored in the res variable. Next, we need to read the http body
	// into a byte array for parsing/processing (using golangs ioutil library):

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return output, err
	}

	// Unmarshal our response to a json type
	var unmarshalledresponse = new(ShortURLResponse)
	err = json.Unmarshal(body, &unmarshalledresponse)
	if err != nil {
		return output, err
	}

	if unmarshalledresponse.Short["destination"] == input {
		return output, errors.New("Could not unshorten")
	}

	output = unmarshalledresponse.Short["destination"]
	fmt.Println()
	return output, nil

}

// GetMoon function
func (h *UtilitiesHandler) GetMoon() (output string, err error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://wttr.in/Moon", nil)
	if err != nil {
		return output, err
	}

	req.Header.Set("User-Agent", "curl/1.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return output, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return output, err
	}

	output = string(body)

	output = vtclean.Clean(output, false)
	payload := strings.Split(output, "\n")

	output = "\n```\n"
	for i, line := range payload {
		if i < 25 {
			fmt.Println(line)
			output = output + line + "\n"
		}
	}
	output = output + "\n```"
	return output, nil

}

// GetCountdownStatus function returns a formatted list of days important to Alpha and pledging
func (h *UtilitiesHandler) GetCountdownStatus() (output string) {

	endofsep := time.Date(2017, 9, 30, 0, 0, 0, 0, time.UTC)
	beginsep := time.Date(2017, 9, 1, 0, 0, 0, 0, time.UTC)
	endofpledges := time.Date(2017, 9, 7, 0, 0, 0, 0, time.UTC)

	exacttimeuntilbegin := beginsep.Sub(time.Now().UTC())
	exattimeuntilend := endofsep.Sub(time.Now().UTC())
	exacttimeuntilpledges := endofpledges.Sub(time.Now().UTC())

	daysuntilbegin := beginsep.YearDay() - time.Now().YearDay()
	daysuntilend := endofsep.YearDay() - time.Now().YearDay()
	daysuntilpledges := endofpledges.YearDay() - time.Now().YearDay()

	output = "Current Important Countdowns: ```\n"
	output = output + "Minimum Estimated Time Until Alpha: " + strconv.Itoa(daysuntilbegin) + " days (Approx: " + TruncateTime(exacttimeuntilbegin, time.Second).String() + ")\n"
	output = output + "Maximum Estimated Time Until Alpha: " + strconv.Itoa(daysuntilend) + " days (Approx: " + TruncateTime(exattimeuntilend, time.Second).String() + ")\n"
	output = output + "Time Until Founders Pack Pledging Ends: " + strconv.Itoa(daysuntilpledges) + " days (Approx: " + TruncateTime(exacttimeuntilpledges, time.Second).String() + ")\n"
	output = output + "```\n"

	return output

}

// GetPledgingStatus function
func (h *UtilitiesHandler) GetPledgingStatus() (output string) {

	endofpledges := time.Date(2017, 9, 7, 0, 0, 0, 0, time.UTC)
	exacttimeuntilpledges := endofpledges.Sub(time.Now().UTC())
	daysuntilpledges := endofpledges.YearDay() - time.Now().YearDay()

	output = "Current Pledging Information: ```\n"
	output = output + "Time Until Founders Pack Pledging Ends: " + strconv.Itoa(daysuntilpledges) + " days (Exactly: " + TruncateTime(exacttimeuntilpledges, time.Second).String() + ")\n"
	output = output + "```\n"

	return output

}

// TayneResponse function
func (h *UtilitiesHandler) TayneResponse(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

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
func (h *UtilitiesHandler) TayneOhGod(url string, s *discordgo.Session, m *discordgo.MessageCreate) {

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
