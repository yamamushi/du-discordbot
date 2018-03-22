package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

type StatsHandler struct {

	registry   	*CommandRegistry
	db         	*DBHandler
	statsdb 	*Stats
	conf 		*Config
}

func (h *StatsHandler) Init(){

}

// RegisterCommands function
func (h *StatsHandler) RegisterCommands() (err error) {
	h.registry.Register("stats", "Manage discord statistics metrics", "")
	return nil
}


func (h *StatsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate){

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"stats") {
		if h.registry.CheckPermission("stats", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Moderator {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

func (h *StatsHandler) ParseCommand(commandarray []string, s *discordgo.Session, m *discordgo.MessageCreate){

	commandarray = RemoveStringFromSlice(commandarray, commandarray[0])
	command, payload := SplitPayload(commandarray)

	if len(payload) < 1 || command == "help" {
		s.ChannelMessageSend(m.ChannelID, "The stats command is used to retrieve various statistics about " +
			"the discord server. You may manually load stats, unload stats, display graphs and various metrics about" +
				" the users on this discord with this command.")
		return
	}

}