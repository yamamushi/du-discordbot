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

// GetAlphaStatus function returns a formatted list of days important to Alpha and pledging
func (h *UtilitiesHandler) GetCountdownStatus() (output string) {

	endofsep := time.Date(2017, 9, 30, 0, 0, 0, 0, time.Now().Location())
	beginsep := time.Date(2017, 9, 1, 0, 0, 0, 0, time.Now().Location())
	endofpledges := time.Date(2017, 9, 7, 0, 0, 0, 0, time.Now().Location())

	daysuntilbegin := beginsep.YearDay() - time.Now().YearDay()
	daysuntilend := endofsep.YearDay() - time.Now().YearDay()
	daysuntilpledges := endofpledges.YearDay() - time.Now().YearDay()

	output = "Current Important Countdowns: ```\n"
	output = output + "Minimum Estimated Days Until Alpha: " + strconv.Itoa(daysuntilbegin) + "\n"
	output = output + "Maximum Estimated Days Until Alpha: " + strconv.Itoa(daysuntilend) + "\n"
	output = output + "Days Until Founders Pack Pledging Ends: " + strconv.Itoa(daysuntilpledges) + "\n"
	output = output + "```\n"

	return output

}

func (h *UtilitiesHandler) GetPledgingStatus() (output string) {

	endofpledges := time.Date(2017, 9, 7, 0, 0, 0, 0, time.Now().Location())
	daysuntilpledges := endofpledges.YearDay() - time.Now().YearDay()
	output = "Current Pledging Information: ```\n"
	output = output + "Days Until Founders Pack Pledging Ends: " + strconv.Itoa(daysuntilpledges) + "\n"
	output = output + "```\n"

	return output

}
