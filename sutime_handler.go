package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"strconv"
	"errors"
	"time"
)

// RecruitmentHandler struct
type SUTimeHandler struct {
	conf     *Config
	registry *CommandRegistry
	db       *DBHandler
	userdb   *UserHandler
}


// Init function
func (h *SUTimeHandler) Init() {
	h.RegisterCommands()
}


// RegisterCommands function
func (h *SUTimeHandler) RegisterCommands() (err error) {
	h.registry.Register("sutime", "Estimate SU Travel Time", "sutime")
	return nil
}

// Read function
func (h *SUTimeHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"sutime") {
		if h.registry.CheckPermission("sutime", m.ChannelID, user) {

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
func (h *SUTimeHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, command + " expects two arguments: <su> <speed>")
		return
	}
	if len(payload) < 2 {
		s.ChannelMessageSend(m.ChannelID, command + " expects two arguments: <su> <speed>")
		return
	}
	estimate, err := h.SUToMinutes(payload[0], payload[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Estimated travel time: " + estimate)
	return
}

func (h *SUTimeHandler) SUToMinutes(distance string, speed string) (conversion string, err error){

	distanceFloat, err := strconv.ParseFloat(distance, 64)
	if err != nil {
		return "", err
	}
	if distanceFloat > 100000000 || distanceFloat <= 0 {
		return "", errors.New("Distance value out of bounds")
	}

	speedFloat := 0.0
	if speed == "max" {
		speedFloat = 30000
	} else {
		speedFloat, err = strconv.ParseFloat(speed, 64)
		if err != nil {
			return "", err
		}
	}

	if speedFloat > 100000 || speedFloat <= 0 {
		return "", errors.New("Speed value out of bounds")
	}
	distanceFloat = distanceFloat * 200.00
	secondsInt := (distanceFloat / speedFloat) * 3600.00

	duration := time.Duration(time.Second * time.Duration(secondsInt))
	return duration.String(), nil
}