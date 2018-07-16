package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"math/rand"
	)

// RecruitmentHandler struct
type RecruitmentHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	recruitmentdb  *RecruitmentDB
	recruitmentChannel string
}


// Init function
func (h *RecruitmentHandler) Init() {
	h.RegisterCommands()
	h.recruitmentdb = &RecruitmentDB{db: h.db}
	h.recruitmentChannel = h.conf.Recruitment.RecruitmentChannel
}


// RegisterCommands function
func (h *RecruitmentHandler) RegisterCommands() (err error) {
	h.registry.Register("recruitment", "Create and Manage Recruitment Ads", "new|edit|delete|info|debug")
	return nil
}

// Read function
func (h *RecruitmentHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"recruitment") {
		if h.registry.CheckPermission("recruitment", m.ChannelID, user) {

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
func (h *RecruitmentHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command " + command + " expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "new" {
		_, commandpayload := SplitPayload(payload)
		h.NewRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "edit" {
		_, commandpayload := SplitPayload(payload)
		h.EditRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "delete" {
		_, commandpayload := SplitPayload(payload)
		h.DeleteRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "info" {
		_, commandpayload := SplitPayload(payload)
		h.RecruitmentInfo(commandpayload, s, m)
		return
	}
	if payload[0] == "debug" {
		_, commandpayload := SplitPayload(payload)
		h.DebugRecruitment(commandpayload, s, m)
		return
	}
}



// HelpOutput function
func (h *RecruitmentHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate){
	output := "Command usage for recruitment: \n"
	output = output + "```\n"
	output = output + "new: Create a new recruitment advertisement\n"
	output = output + "edit: update an existing recruitment ad\n"
	output = output + "delete: delete a recruitment advertisement\n"
	output = output + "info: display information about a recruitment advertisement\n"
	output = output + "debug: an admin command for retrieving recruitment ad debug information\n"
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}

func (h *RecruitmentHandler) NewRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

}

func (h *RecruitmentHandler) EditRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

}

func (h *RecruitmentHandler) DeleteRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

}

func (h *RecruitmentHandler) RecruitmentInfo(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

}

func (h *RecruitmentHandler) DebugRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

}

func (h *RecruitmentHandler) ShuffleRecords(DisplayRecords []RecruitmentDisplayRecord) (ShuffledRecords []RecruitmentDisplayRecord){

	for i := len(DisplayRecords) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		DisplayRecords[i], DisplayRecords[j] = DisplayRecords[j], DisplayRecords[i]
	}

	return DisplayRecords
}