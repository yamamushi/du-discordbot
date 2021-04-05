package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
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

	if len(payload) == 0 || len(payload) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, command+" expects two arguments: <su> <speed>")
		return
	}

	speedFloat := 0.0
	payload[1] = strings.ToLower(payload[1])
	speedFloat, err := strconv.ParseFloat(payload[1], 64)
	if err != nil {
		if strings.HasSuffix(payload[1], "k") {
			payload[1] = strings.TrimSuffix(payload[1], "k")
			speedFloat, err = strconv.ParseFloat(payload[1], 64)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
				return
			}
			speedFloat = speedFloat * 1000
		} else if strings.ToLower(payload[1]) == "max" {
			speedFloat = 30000
		}
	}

	distanceFloat, err := strconv.ParseFloat(payload[0], 64)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	estimate, err := h.CalculateTime(distanceFloat, speedFloat)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "Estimated travel time: "+estimate)
	return
}

func (h *SUTimeHandler) CalculateTime(distanceFloat float64, speedFloat float64) (duration string, err error) {

	if distanceFloat > 100000000 || distanceFloat <= 0 {
		return "", errors.New("Distance value out of bounds")
	}

	if speedFloat > 100000 || speedFloat <= 0 {
		return "", errors.New("Speed value out of bounds")
	}
	distanceFloat = distanceFloat * 200.00
	secondsInt := (distanceFloat / speedFloat) * 3600.00

	timeDuration := time.Duration(time.Second * time.Duration(secondsInt))
	return timeDuration.String(), nil

}
