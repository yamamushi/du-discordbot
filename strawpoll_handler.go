package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

type StrawpollHandler struct {
	conf     *Config
	registry *CommandRegistry
	db       *DBHandler
	API      string
}

type StrawpollPost struct {
	Title       string      `json:"title"`
	Options     []string    `json:"options"`
	Multi       bool        `json:"multi"`
	Dupcheck    string      `json:"dupcheck"`
	Captcha     bool        `json:"captcha"`
}

type StrawpollPostResponse struct {
	ID          int      `json:"id"`
	Title       string      `json:"title"`
	Options     []string    `json:"options"`
	Multi       bool        `json:"multi"`
	Dupcheck    string      `json:"dupcheck"`
	Captcha     bool        `json:"captcha"`
}


// Init function
func (h *StrawpollHandler) Init() {
	h.RegisterCommands()
	h.API = h.conf.APIConfig.Strawpoll
}


// RegisterCommands function
func (h *StrawpollHandler) RegisterCommands() (err error) {
	h.registry.Register("strawpoll", "Create a strawpoll", h.conf.DUBotConfig.CP+"strawpoll title | option 1 | option 2")
	return nil
}

// Read function
func (h *StrawpollHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"strawpoll") {
		if h.registry.CheckPermission("strawpoll", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				//fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Citizen {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

func (h *StrawpollHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	_, payload := SplitCommandFromArgs(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: ```"+h.conf.DUBotConfig.CP+"strawpoll title | option 1 | option 2 \n```\n")
		return
	}

	title, options := h.splitOptions(payload)
	if len(options) < 2 {
		s.ChannelMessageSend(m.ChannelID, "You must provide at least two options!")
		return
	}

	requestString := &StrawpollPost{Title:title, Options: options, Multi: false, Dupcheck: "normal", Captcha: true}
	byteString, err := json.Marshal(requestString)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error Marshall: " + err.Error())
		return
	}

	req, err := http.NewRequest("POST", h.conf.APIConfig.Strawpoll, bytes.NewBuffer(byteString))
	req.Header.Set("X-Custom-Header", "DualUniverseDiscordBot")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error Post: " + err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	responseJson := StrawpollPostResponse{}
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error Unmarshall: " + err.Error())
		return
	}
	link := "https://www.strawpoll.me/" + strconv.Itoa(responseJson.ID)
	s.ChannelMessageSend(m.ChannelID, ":ballot_box: | **"+m.Author.Username+"'s Poll** \n" + link)
	return
}



func (h *StrawpollHandler) splitOptions(message string) (title string, options []string) {

	split := strings.Split(message, "|")

	for i, word := range split {
		split[i] = strings.TrimPrefix(word, " ")
		if i > 0 {
			options = append(options, split[i])
		} else {
			title = split[i]
		}
	}


	return title, options
}
