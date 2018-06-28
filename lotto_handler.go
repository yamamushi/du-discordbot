package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"time"
	"encoding/json"
	"strconv"
	"errors"
)

// NotificationsHandler struct
type LottoHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	lottodb  *LottoDB
}

type NewGiveaway struct {
	Shortname string
	Description string
	Duration string
}

// Init function
func (h *LottoHandler) Init() {
	h.RegisterCommands()
	h.lottodb = &LottoDB{db: h.db}
}


// RegisterCommands function
func (h *LottoHandler) RegisterCommands() (err error) {
	h.registry.Register("giveaway", "Manage giveaways", "enter|new|update|end|list|history")
	return nil
}

// Read function
func (h *LottoHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"giveaway") {
		if h.registry.CheckPermission("giveaway", m.ChannelID, user) {

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



// ParseCommand function
func (h *LottoHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command " + command + " expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "enter" {
		_, commandpayload := SplitPayload(payload)
		h.EnterGiveaway(commandpayload, s, m)
		return
	}
	if payload[0] == "update" {
		_, commandpayload := SplitPayload(payload)
		h.UpdateGiveaway(commandpayload, s, m)
		return
	}
	if payload[0] == "new" {
		_, commandpayload := SplitPayload(payload)
		h.NewGiveaway(commandpayload, s, m)
		return
	}
	if payload[0] == "end" {
		_, commandpayload := SplitPayload(payload)
		h.EndGiveaway(commandpayload, s, m)
		return
	}
	if payload[0] == "list" {
		_, commandpayload := SplitPayload(payload)
		h.ListGiveaways(commandpayload, s, m)
		return
	}
	if payload[0] == "history" {
		_, commandpayload := SplitPayload(payload)
		h.GiveawayHistory(commandpayload, s, m)
		return
	}
}

// HelpOutput function
func (h *LottoHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate){
	output := "Command usage for giveaway: \n"
	output = output + "```\n"
	output = output + "enter: enter a giveaway by name or ID\n"
	output = output + "new: start a new giveaway\n"
	output = output + "update: update an existing giveaway\n"
	output = output + "end: end a giveaway\n"
	output = output + "list: list currently active giveaways\n"
	output = output + "history: list x-giveaways and their winners\n"
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}

func (h *LottoHandler) NewGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'new' expects a formatted giveaway, see help for usage.")
		return
	}

	if strings.ToLower(payload[0]) == "help" {

		examplePayload := "{\n\t\"shortname\": \"IWannaWin\",\n\t\"description\": \"Enter to win a prize!\",\n\t\"duration\" : \"1d 2h 3m\"\n}"
		s.ChannelMessageSend(m.ChannelID, "'new' expects a payload formatted in json. Example: ```" +examplePayload+ "\n```" )
		return
	}

	var combined string
	for count, i := range payload {
		if count != 0 && count != len(payload) - 1 {
			combined += i + " "
		}
	}
	fmt.Println(combined)
	unpacked, err := h.UnpackGiveaway(combined)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error unpacking payload: " + err.Error())
		return
	}

	if strings.Contains(unpacked.Shortname, " ") ||  strings.Contains(unpacked.Shortname, "\n") {
		s.ChannelMessageSend(m.ChannelID, "Shortname cannot contain spaces!")
		return
	}

	fmt.Println("Duration: " + unpacked.Duration)
	days, hours, minutes, err := h.ParseDuration(unpacked.Duration)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing duration: " + err.Error())
		return
	}
	if days > 0 {
		minutes = (24 * 60) + minutes
	}
	if hours > 0 {
		minutes = (hours * 60) + minutes
	}

	duration := time.Duration(minutes * 60 * 1000 * 1000 * 1000)
	//fmt.Println("Interval: " + strconv.Itoa(interval))

	record, err := h.CreateGiveaway(m.Author.ID, unpacked.Shortname, unpacked.Description, duration)
	err = h.lottodb.AddLottoRecordToDB(record)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not save record to database: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Payload unpacked")
	return
}


func (h *LottoHandler) ParseDuration(duration string) (days int64, hours int64, minutes int64, err error) {

	daysstring := "0"
	hoursstring := "0"
	minutesstring := "0"

	if !strings.Contains(duration, " ") {
		return 0, 0, 0, errors.New("Invalid time interval format")
	}

	separated := strings.Split(duration, " ")

	for _, field := range separated {

		for _, value := range field {
			switch {
			case value >= '0' && value <= '9':
				if strings.Contains(field, "d") {
					daysstring = strings.TrimSuffix(field, "d")
					hours, err = strconv.ParseInt(daysstring, 10, 64)
					if err != nil {
						return 0, 0, 0, errors.New("Could not parse hours")
					}
				} else if strings.Contains(field, "h") {
					hoursstring = strings.TrimSuffix(field, "h")
					hours, err = strconv.ParseInt(hoursstring, 10, 64)
					if err != nil {
						return 0, 0, 0, errors.New("Could not parse hours")
					}
				} else if strings.Contains(field, "m") {
					minutesstring = strings.TrimSuffix(field, "m")
					minutes, err = strconv.ParseInt(minutesstring, 10, 64)
					if err != nil {
						return 0, 0, 0, errors.New("Could not parse minutes")
					}
				} else {
					return 0, 0, 0, errors.New("Invalid time interval format")
				}
				break
			default:
				return 0, 0, 0, errors.New("Invalid time interval format")
			}
			break
		}
	}

	if days == 0 && hours == 0 && minutes == 0 {
		return days, hours, minutes, errors.New("Invalid interval specified")
	}

	return days, hours, minutes, nil
}


func (h *LottoHandler) UnpackGiveaway(payload string) (unpacked NewGiveaway, err error) {

	payload = strings.TrimPrefix(payload, "~giveaway new ") // This all will need to be updated later, this is just
	payload = strings.TrimPrefix(payload, "\n")             // A lazy way of cleaning the command
	payload = strings.TrimPrefix(payload, "```")
	payload = strings.TrimSuffix(payload, "```")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.Trim(payload, "```")

	unmarshallcontainer := NewGiveaway{}
	if err := json.Unmarshal([]byte(payload), &unmarshallcontainer); err != nil {
		return NewGiveaway{}, err
	} else {
		return unmarshallcontainer, nil
	}
}

func (h *LottoHandler) CreateGiveaway(ownerID string, shortname string, description string, duration time.Duration) (record LottoRecord, err error) {

	record.ID, err = GetUUID()
	if err != nil {
		return record, err
	}
	record.OwnerID = ownerID
	record.ShortName = shortname
	record.Description = description
	record.CreatedDate = time.Now()
	record.Duration = duration
	record.Completed = false

	return record, nil
}

func (h *LottoHandler) EnterGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'enter' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}


func (h *LottoHandler) UpdateGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'update' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *LottoHandler) EndGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'end' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *LottoHandler) ListGiveaways(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){

	records, err := h.lottodb.GetAllLottoDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading database: " + err.Error())
		return
	}
	for _, record := range records {
		fmt.Println(record.ShortName)
	}

	s.ChannelMessageSend(m.ChannelID, "Under construction!")
	return
}

func (h *LottoHandler) GiveawayHistory(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'history' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}