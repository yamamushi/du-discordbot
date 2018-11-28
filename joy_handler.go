package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type JoyHandler struct {
	conf     *Config
	registry *CommandRegistry
	db       *DBHandler
	userdb   *UserHandler
	joydb    *JoyDB
}

// Init function
func (h *JoyHandler) Init() {
	h.RegisterCommands()
	h.joydb = &JoyDB{db: h.db}
}


// RegisterCommands function
func (h *JoyHandler) RegisterCommands() (err error) {
	h.registry.Register("joy", "Bring :joy: to the world", "joy")
	return nil
}

// Read function
func (h *JoyHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"joy") {
		if h.registry.CheckPermission("joy", m.ChannelID, user) {

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

func (h *JoyHandler) React( s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	//message := strings.ToLower(m.Content)

	// Check author in DB
	if !h.CheckUser(m.Author.ID) {
		return
	}
	// Check message has joy
	if strings.Contains(m.Content, "😂"){
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, "😂")
		return
	}
	return
}

// ParseCommand function
func (h *JoyHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(command) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'joy' command, see usage for more info")
		return
	}

	if command[1] == "enable" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "enable" && len(command) > 2 {
		h.EnableJoyUser(command, s, m)
		return
	}

	if command[1] == "disable" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "disable" && len(command) > 2 {
		h.DisableJoyUser(command, s, m)
		return
	}
}

func (h *JoyHandler) EnableJoyUser( command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "This command requires a user mention")
		return
	}

	record, err := h.joydb.GetJoyIDFromDB(m.Mentions[0].ID)
	if err != nil {
		record.UserID = m.Mentions[0].ID
		record.Enabled = true
		err = h.joydb.AddJoyRecordToDB(record)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Enabled for " + m.Mentions[0].Mention())
		return
	}

	record.Enabled = true
	h.joydb.UpdateJoyRecord(record)
	s.ChannelMessageSend(m.ChannelID, "Enabled for " + m.Mentions[0].Mention())
	return
}

func (h *JoyHandler) DisableJoyUser( command []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "This command requires a user mention")
		return
	}

	record, err := h.joydb.GetJoyIDFromDB(m.Mentions[0].ID)
	if err != nil {
		record.UserID = m.Mentions[0].ID
		record.Enabled = false
		err = h.joydb.AddJoyRecordToDB(record)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Disabled for " + m.Mentions[0].Mention())
		return
	}

	record.Enabled = false
	h.joydb.UpdateJoyRecord(record)
	s.ChannelMessageSend(m.ChannelID, "Disabled for " + m.Mentions[0].Mention())
	return
}

func (h *JoyHandler) CheckUser( userID string ) bool {
	record, err := h.joydb.GetJoyIDFromDB(userID)
	if err != nil {
		return false
	} else {
		if record.Enabled {
			return true
		}
	}
	return false
}