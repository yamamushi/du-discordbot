package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
		"sync"
	"time"
	"strconv"
)

// AdminHandler struct
type AdminHandler struct {
	conf *Config
	db   *DBHandler

	registry *CommandRegistry
	reactions *ReactionsHandler
	querylocker  sync.RWMutex
	configdb     *ConfigDB
	userdb    *UserHandler
	globalstate *StateDB
}


// Init function
func (h *AdminHandler) Init() {
	h.RegisterCommands()
}

// RegisterCommands function
func (h *AdminHandler) RegisterCommands() (err error) {

	h.registry.Register("admin", "Open the administrative interface", "-")
	return nil

}


func (h *AdminHandler) Flush(s *discordgo.Session, m *discordgo.MessageCreate) {
	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	if strings.HasPrefix(m.Content, cp+"flush") {
		// Grab our sender ID to verify if this user has permission to use this command
		db := h.db.rawdb.From("Users")
		var user User
		err := db.One("ID", m.Author.ID, &user)
		if err != nil {
			fmt.Println("error retrieving user:" + m.Author.ID)
		}

		if user.Moderator {
			message := strings.Fields(m.Content)
			if len(message) < 2 {
				response, err := s.ChannelMessageSend(m.ChannelID, ":rotating_light: Expected a value for flush!")
				if err == nil {
					time.Sleep(3*time.Second)
					s.ChannelMessageDelete(m.ChannelID, response.ID)
					s.ChannelMessageDelete(m.ChannelID, m.ID)
				}
				return
			}
			h.FlushChannel(message[1], s, m)
			return
		}
	}
}

// FlushChannel function
func (h *AdminHandler) FlushChannel(amount string, s *discordgo.Session, m *discordgo.MessageCreate) {

	_, err := s.Channel(m.ChannelID)
	if err != nil {
		response, err := s.ChannelMessageSend(m.ChannelID, ":rotating_light: Error getting channel: "+err.Error())
		if err == nil {
			time.Sleep(3*time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}
		return
	}

	count, err := strconv.Atoi(amount)
	if err != nil {
		response, err := s.ChannelMessageSend(m.ChannelID, ":rotating_light: Error with count: "+err.Error())
		if err == nil {
			time.Sleep(3*time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}
		return
	}

	err = FlushMessages(s, m.ChannelID, count+1) // +1 to account for the flush command itself
	if err != nil {
		response, err := s.ChannelMessageSend(m.ChannelID, ":rotating_light: Error flushing channel: "+err.Error())
		if err == nil {
			time.Sleep(3*time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}
		return
	}

	validated, err := s.ChannelMessageSend(m.ChannelID, ":ballot_box_with_check:  Deleted "+amount+" messages from channel!")
	if err != nil {
		//s.ChannelMessageSend(m.ChannelID, ":rotating_light: Error flushing channel: "+err.Error())
		return
	}
	sleeptime := time.Duration(time.Second * 3)
	time.Sleep(sleeptime)

	err = s.ChannelMessageDelete(m.ChannelID, validated.ID)
	if err != nil {
		return
	}
	return
}


// Read function
func (h *AdminHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"admin") {
		if h.registry.CheckPermission("admin", m.ChannelID, user) {

			command := strings.Fields(m.Content)
			if command[0] != cp+"admin" {
				return
			}

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

func (h *AdminHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	_, payload := SplitPayload(commandlist)

	if len(payload) == 1 {
		if payload[0] == "help" {
			h.HelpOutput(s, m)
			return
		}
	}

	mainMenu := h.MainMenu()
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, mainMenu)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}



func (h *AdminHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "The administrative interface is managed through a ui. Click on the reaction emojis to navigate.")
	return
}


func (h *AdminHandler) DefaultEmbed() (embed *discordgo.MessageEmbed) {

	embed = &discordgo.MessageEmbed{}

	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: "https://cdn.discordapp.com/attachments/418457755276410880/473080359219625989/Server_Logo.jpg"}
	embed.Color = 16265732

	//loc, _ := time.LoadLocation("America/Chicago")
	//embed.Timestamp = time.Now().In(loc).Format("Mon Jan _2 03:04 MST 2006")

	embed.Footer = &discordgo.MessageEmbedFooter{Text:"Dual Universe Bot",
	IconURL:"https://cdn.discordapp.com/attachments/418457755276410880/473080359219625989/Server_Logo.jpg"}

	embed.Author = &discordgo.MessageEmbedAuthor{Name:"Admin Interface"}

	return embed
}

func (h *AdminHandler) MainMenu() (embed *discordgo.MessageEmbed){
	embed = h.DefaultEmbed()

	embed.Title = "Main Menu"

	loc, _ := time.LoadLocation("America/Chicago")
	//embed.Timestamp = time.Now().In(loc).Format("Mon Jan _2 03:04 MST 2006")
	embed.Description = "Welcome to the Dual Universe Bot Admin Interface. Please select an option from the menu below.\n\n" + time.Now().In(loc).Format("Mon Jan _2 03:04 MST 2006")

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

	var fields []*discordgo.MessageEmbedField

	configs := &discordgo.MessageEmbedField{Name: "0⃣", Value: "Configs", Inline: true}
	fields = append(fields, configs)




	embed.Fields = fields
	return embed
}











