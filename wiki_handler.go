package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type WikiHandler struct {
	conf      *Config
	registry  *CommandRegistry
	db        *DBHandler
	callback  *CallbackHandler
	reactions *ReactionsHandler

	userdb   *UserHandler
	configdb *ConfigDB
}

type WikiSearchResult struct {
	Batchcomplete string `json:"batchcomplete"`
	Continue      struct {
		Sroffset int    `json:"sroffset"`
		Continue string `json:"continue"`
	} `json:"continue"`
	Query struct {
		Searchinfo struct {
			Totalhits int `json:"totalhits"`
		} `json:"searchinfo"`
		Search []struct {
			Ns        int       `json:"ns"`
			Title     string    `json:"title"`
			Size      int       `json:"size"`
			Wordcount int       `json:"wordcount"`
			Snippet   string    `json:"snippet"`
			Timestamp time.Time `json:"timestamp"`
		} `json:"search"`
	} `json:"query"`
}

// Init function
func (h *WikiHandler) Init() {
	h.RegisterCommands()
}

// RegisterCommands function
func (h *WikiHandler) RegisterCommands() (err error) {
	h.registry.Register("wiki", "Interact with the Dual Universe wiki", "<search term>")
	return nil
}

// Read function
func (h *WikiHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"wiki") {
		if h.registry.CheckPermission("wiki", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Citizen {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// ParseCommand function
func (h *WikiHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command "+command+" expects an argument, see help for usage.")
		return
	}
	if payload[0] == "!help" {
		h.HelpOutput(s, m)
		return
	}

	searchquery := ""
	for i, word := range payload {
		if i == 0 {
			searchquery = word
		} else {
			searchquery = searchquery + "+" + word
		}
	}

	h.ParseSearch(searchquery, s, m)
	return
}

func (h *WikiHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate) {
	output := "Command usage for wiki: \n"
	output = output + "```\n"
	output = output + "This command interacts with the Official Dual Universe wiki at https://dualuniverse.gamepedia.com/" +
		"\n\nInteract with the menus provided to search for information."
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}

func (h *WikiHandler) ParseSearch(content string, s *discordgo.Session, m *discordgo.MessageCreate) {

	jsonresponse, err := h.GetQueryJson(content, 0)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if len(jsonresponse.Query.Search) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Sorry, no search results were found.")
		return
	}

	if len(jsonresponse.Query.Search) == 1 {
		output := ":bulb: " + "https://dualuniverse.gamepedia.com/" + strings.Replace(jsonresponse.Query.Search[0].Title, " ", "_", -1)
		s.ChannelMessageSend(m.ChannelID, output)
		return
	}

	var reactions []string

	output := ":bulb: Pages with content matching your query: \n```\n"
	//output = output + "(Page " + strconv.Itoa(1) + ")\n"
	for i, searchresult := range jsonresponse.Query.Search {
		if i < 10 {
			output = output + strconv.Itoa(i) + ") " + searchresult.Title + "\n"
		}
		if i == 0 {
			reactions = append(reactions, "0⃣")
		}
		if i == 1 {
			reactions = append(reactions, "1⃣")
		}
		if i == 2 {
			reactions = append(reactions, "2⃣")
		}
		if i == 3 {
			reactions = append(reactions, "3⃣")
		}
		if i == 4 {
			reactions = append(reactions, "4⃣")
		}
		if i == 5 {
			reactions = append(reactions, "5⃣")
		}
		if i == 6 {
			reactions = append(reactions, "6⃣")
		}
		if i == 7 {
			reactions = append(reactions, "7⃣")
		}
		if i == 8 {
			reactions = append(reactions, "8⃣")
		}
		if i == 9 {
			reactions = append(reactions, "9⃣")
			reactions = append(reactions, "➡")
		}
	}
	output = output + "\n```\n"

	marshalled, err := json.Marshal(jsonresponse)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	packed := content + "||" + string(marshalled)

	err = h.reactions.Create(h.HandlePendingCreatedReaction, reactions, m.ChannelID, output, packed, s)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}
}

func (h *WikiHandler) GetQueryJson(search string, offset int) (result WikiSearchResult, err error) {

	req, err := http.NewRequest("GET", "https://dualuniverse.gamepedia.com//api.php?action=query&format=json&list=search&srsearch="+search+"&srlimit=10h&sroffset="+strconv.Itoa(offset), nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("User-Agent", "dualuniverse-discord/1.0.0")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (h *WikiHandler) HandlePendingCreatedReaction(reaction string, payload string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()

	splitstring := strings.Split(payload, "||")
	query := splitstring[0]

	combined := ""
	for i, word := range splitstring {
		if i > 0 {
			combined = combined + word
		}
	}

	unmarshalledjson := WikiSearchResult{}
	err := json.Unmarshal([]byte(combined), &unmarshalledjson)
	if err != nil {
		fmt.Println(combined)
		s.ChannelMessageSend(channelID, "Error: "+err.Error())
		return
	}

	if reaction == "➡" {
		jsonresponse, err := h.GetQueryJson(query, unmarshalledjson.Continue.Sroffset+10)
		if err != nil {
			s.ChannelMessageSend(channelID, "Error: "+err.Error())
			return
		}

		output := ":bulb: Pages with content matching your query: \n```\n"
		//output = output + "(Page " + strconv.Itoa(10/jsonresponse.Continue.Sroffset) + ")\n"
		for i, searchresult := range jsonresponse.Query.Search {
			if i < 10 {
				output = output + strconv.Itoa(i) + ") " + searchresult.Title + "\n"
			}
		}
		output = output + "\n```\n"
		sentmsg, err := s.ChannelMessageEdit(channelID, messageID, output)
		if err != nil {
			fmt.Println(sentmsg.ID + " " + err.Error())
		}
		err = s.MessageReactionAdd(channelID, messageID, "⬅")
		if err != nil {
			s.ChannelMessageSend(channelID, "Error: "+err.Error())
			return
		}
		return
	}
	if reaction == "⬅" {
		if unmarshalledjson.Continue.Sroffset == 0 {
			return
		}

		jsonresponse, err := h.GetQueryJson(query, unmarshalledjson.Continue.Sroffset-10)
		if err != nil {
			s.ChannelMessageSend(channelID, "Error: "+err.Error())
			return
		}

		output := ":bulb: Pages with content matching your query: \n```\n"
		//output = output + "(Page " + strconv.Itoa(jsonresponse.Continue.Sroffset) + ")\n"
		for i, searchresult := range jsonresponse.Query.Search {
			if i < 10 {
				output = output + strconv.Itoa(i) + ") " + searchresult.Title + "\n"
			}
		}
		output = output + "\n```\n"
		s.ChannelMessageEdit(channelID, messageID, output)
		return
	}
	if reaction == "0⃣" {
		h.SendResult(unmarshalledjson, 0, channelID, messageID, s)
	}
	if reaction == "1⃣" {
		h.SendResult(unmarshalledjson, 1, channelID, messageID, s)
	}
	if reaction == "2⃣" {
		h.SendResult(unmarshalledjson, 2, channelID, messageID, s)
	}
	if reaction == "3⃣" {
		h.SendResult(unmarshalledjson, 3, channelID, messageID, s)
	}
	if reaction == "4⃣" {
		h.SendResult(unmarshalledjson, 4, channelID, messageID, s)
	}
	if reaction == "5⃣" {
		h.SendResult(unmarshalledjson, 5, channelID, messageID, s)
	}
	if reaction == "6⃣" {
		h.SendResult(unmarshalledjson, 6, channelID, messageID, s)
	}
	if reaction == "7⃣" {
		h.SendResult(unmarshalledjson, 7, channelID, messageID, s)
	}
	if reaction == "8⃣" {
		h.SendResult(unmarshalledjson, 8, channelID, messageID, s)
	}
	if reaction == "9⃣" {
		h.SendResult(unmarshalledjson, 9, channelID, messageID, s)
	}

	return
}

func (h *WikiHandler) SendResult(result WikiSearchResult, selection int, channelID string, messageID string, s *discordgo.Session) {

	if selection > len(result.Query.Search) {
		return
	}

	/*
		output := ":bulb: "+result.Query.Search[selection].Title+" : \n```\n"
		//output = output + "(Page " + strconv.Itoa(jsonresponse.Continue.Sroffset) + ")\n"

		output = output + strip.StripTags(result.Query.Search[selection].Snippet) + "...\n"
		output = output + "\n"
		output = output + "Read more at: https://dualuniverse.gamepedia.com/"+strings.Replace(result.Query.Search[selection].Title, " ", "%20", -1) + "\n"

		loc, _ := time.LoadLocation("America/Chicago")
		output = output + result.Query.Search[selection].Timestamp.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"

		output = output + "\n```\n"
	*/
	output := ":bulb: " + "https://dualuniverse.gamepedia.com/" + strings.Replace(result.Query.Search[selection].Title, " ", "_", -1)
	s.ChannelMessageEdit(channelID, messageID, output)
	h.reactions.UnWatch(channelID, messageID)
	/*
		var reactions []string
		reactions = append(reactions, "0⃣")
		reactions = append(reactions, "1⃣")
		reactions = append(reactions, "2⃣")
		reactions = append(reactions, "3⃣")
		reactions = append(reactions, "4⃣")
		reactions = append(reactions, "5⃣")
		reactions = append(reactions, "6⃣")
		reactions = append(reactions, "7⃣")
		reactions = append(reactions, "8⃣")
		reactions = append(reactions, "9⃣")
		reactions = append(reactions, "➡")
		reactions = append(reactions, "⬅")

		for _, reaction := range reactions {
			s.MessageReactionRemove(channelID, messageID, reaction, s.State.User.ID)
		}
	*/
	//time.Sleep(5*time.Second)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}
