package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// NotificationsHandler struct
type GiveawayHandler struct {
	conf       *Config
	registry   *CommandRegistry
	callback   *CallbackHandler
	db         *DBHandler
	giveawaydb *GiveawayDB
}

type NewGiveaway struct {
	Shortname   string
	Description string
	Duration    string
}

// Init function
func (h *GiveawayHandler) Init() {
	h.RegisterCommands()
	h.giveawaydb = &GiveawayDB{db: h.db}
}

// RegisterCommands function
func (h *GiveawayHandler) RegisterCommands() (err error) {
	h.registry.Register("giveaway", "Manage giveaways", "enter|new|update|end|list|history")
	return nil
}

// Read function
func (h *GiveawayHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			if user.Citizen {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// ParseCommand function
func (h *GiveawayHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command "+command+" expects an argument, see help for usage.")
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
	if payload[0] == "info" {
		_, commandpayload := SplitPayload(payload)
		h.GiveawayInfo(commandpayload, s, m)
		return
	}
	if payload[0] == "history" {
		_, commandpayload := SplitPayload(payload)
		h.GiveawayHistory(commandpayload, s, m)
		return
	}
}

// HelpOutput function
func (h *GiveawayHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func (h *GiveawayHandler) GiveawayWatcher(s *discordgo.Session) {
	for true {
		// Only run every X minutes
		time.Sleep(h.conf.DUBotConfig.GiveawayTimer * time.Minute)

		recordlist, err := h.giveawaydb.GetAllGiveawayDB()
		if err != nil {
			fmt.Println("Error reading from giveaway database: " + err.Error())
		} else {
			for _, record := range recordlist {
				endTime := record.CreatedDate.Add(record.Duration)
				if time.Now().After(endTime) && record.Active {
					winnerID, err := h.PickWinner(record.ID)
					if err == nil {
						s.ChannelMessageSend(h.conf.DUBotConfig.GiveawayChannel, "<@"+winnerID+"> has won the "+record.ShortName+" giveaway! Congratulations! Please contact <@"+record.OwnerID+"> to claim your prize!")

					} else {
						s.ChannelMessageSend(h.conf.DUBotConfig.GiveawayChannel, record.ShortName+" has ended without a winner because no entries were found.")
					}
					h.DeactivateGiveaway(record.ShortName, winnerID)
				}
			}
		}
	}
}

func (h *GiveawayHandler) PickWinner(giveawayid string) (winnerID string, err error) {
	var activeentrylist []GiveawayEntry
	recordlist, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		return "", err
	}
	for _, record := range recordlist {
		//fmt.Println(giveawayid + " , " + record.ID)
		if record.ID == giveawayid {
			entrylist, err := h.giveawaydb.GetAllEntryDB()
			if err != nil {
				return "", err
			}

			for _, entry := range entrylist {
				if entry.GiveawayID == record.ID {
					activeentrylist = append(activeentrylist, entry)
				}
			}
			rand.Seed(time.Now().Unix())
			winnerID = activeentrylist[rand.Intn(len(activeentrylist))].UserID
			err = h.giveawaydb.FlushEntriesForGiveaway(record.ID)
			return winnerID, err
		}
	}
	return "", errors.New("no entries found")
}

func (h *GiveawayHandler) NewGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
	}

	if !user.Admin {
		return
	}

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'new' expects a formatted giveaway, see help for usage.")
		return
	}

	if strings.ToLower(payload[0]) == "help" {

		examplePayload := "{\n\t\"shortname\": \"IWannaWin\",\n\t\"description\": \"Enter to win a prize!\",\n\t\"duration\" : \"1d 2h 3m\"\n}"
		s.ChannelMessageSend(m.ChannelID, "'new' expects a payload formatted in json. Example: ```"+examplePayload+"\n```")
		return
	}

	var combined string
	for count, i := range payload {
		if count != 0 && count != len(payload)-1 {
			combined += i + " "
		}
	}
	//fmt.Println(combined)
	unpacked, err := h.UnpackGiveaway(combined)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error unpacking payload: "+err.Error())
		return
	}

	if strings.Contains(unpacked.Shortname, " ") || strings.Contains(unpacked.Shortname, "\n") {
		s.ChannelMessageSend(m.ChannelID, "Shortname cannot contain spaces!")
		return
	}

	//fmt.Println("Duration: " + unpacked.Duration)
	days, hours, minutes, err := h.ParseDuration(unpacked.Duration)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing duration: "+err.Error())
		return
	}
	minutes = (days * 24 * 60) + (hours * 60) + minutes

	duration := time.Duration(minutes * 60 * 1000 * 1000 * 1000)
	//fmt.Println("Interval: " + strconv.Itoa(interval))

	record, err := h.CreateGiveaway(m.Author.ID, unpacked.Shortname, unpacked.Description, duration)

	giveawayrecords, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not read database: "+err.Error())
		return
	}
	for _, giveawayrecord := range giveawayrecords {
		if giveawayrecord.ShortName == record.ShortName {
			s.ChannelMessageSend(m.ChannelID, "Error: Active Giveaway with short name "+record.ShortName+" already exists.")
			return
		}
	}

	err = h.giveawaydb.AddGiveawayRecordToDB(record)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not save record to database: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Payload unpacked")
	return
}

func (h *GiveawayHandler) ParseDuration(duration string) (days int64, hours int64, minutes int64, err error) {

	daysstring := "0"
	hoursstring := "0"
	minutesstring := "0"

	if !strings.Contains(duration, " ") && len(duration) > 3 {
		return 0, 0, 0, errors.New("Invalid time interval format")
	}

	separated := strings.Split(duration, " ")

	for _, field := range separated {

		for _, value := range field {
			switch {
			case value >= '0' && value <= '9':
				if strings.Contains(field, "d") {
					daysstring = strings.TrimSuffix(field, "d")
					days, err = strconv.ParseInt(daysstring, 10, 64)
					if err != nil {
						return 0, 0, 0, errors.New("Could not parse days")
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

func (h *GiveawayHandler) UnpackGiveaway(payload string) (unpacked NewGiveaway, err error) {

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

func (h *GiveawayHandler) CreateGiveaway(ownerID string, shortname string, description string, duration time.Duration) (record GiveawayRecord, err error) {

	record.ID, err = GetUUID()
	if err != nil {
		return record, err
	}
	record.OwnerID = ownerID
	record.ShortName = shortname
	record.Description = description
	record.CreatedDate = time.Now()
	record.Duration = duration
	record.Active = true

	return record, nil
}

func (h *GiveawayHandler) EnterGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'enter' expects an argument.")
		return
	}

	giveawayrecords, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading database: "+err.Error())
		return
	}

	shortname := strings.ToLower(payload[0])
	for _, giveawayrecord := range giveawayrecords {
		if strings.ToLower(giveawayrecord.ShortName) == strings.ToLower(shortname) {
			entries, err := h.giveawaydb.GetAllEntryDB()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error validating entry: "+err.Error())
				return
			}

			for _, entry := range entries {
				if entry.UserID == m.Author.ID && entry.GiveawayID == giveawayrecord.ID {
					s.ChannelMessageSend(m.ChannelID, "You have already entered this giveaway!")
					return
				}
			}

			if giveawayrecord.Restricted {
				discordmember, err := s.GuildMember(h.conf.DiscordConfig.GuildID, m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error validating user discord record: "+err.Error())
					return
				}

				userroles := discordmember.Roles
				for _, role := range userroles {
					rolename, err := getRoleNameByID(role, h.conf.DiscordConfig.GuildID, s)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Error validating roles: "+err.Error())
						return
					}

					if rolename == "Alpha Authorized" || rolename == "ATV Authorized" || rolename == "Pre Alpha Authorized" {
						s.ChannelMessageSend(m.ChannelID, "You are not allowed to enter this giveaway, as it is only for users who do not currently have Alpha access.")
						return
					}
				}
			}

			newentry := GiveawayEntry{}
			newentry.Date = time.Now()

			newentry.ID, err = GetUUID()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error generating entry: "+err.Error())
				return
			}

			newentry.UserID = m.Author.ID
			newentry.GiveawayID = giveawayrecord.ID

			err = h.giveawaydb.AddEntryRecordToDB(newentry)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error saving entry: "+err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Entry confirmed, good luck!")
			return
		}
	}
	s.ChannelMessageSend(m.ChannelID, "No giveaway with name "+shortname+" exists!")
	return
}

func (h *GiveawayHandler) UpdateGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
	}

	if !user.Admin {
		return
	}
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'update' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *GiveawayHandler) EndGiveaway(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
	}

	if !user.Admin {
		return
	}

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'end' expects an argument.")
		return
	}

	payload[0] = strings.ToLower(payload[0])
	giveawayID := ""
	giveawayownerID := ""
	giveawayrecords, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not read database: "+err.Error())
		return
	}
	for _, giveawayrecord := range giveawayrecords {
		if strings.ToLower(giveawayrecord.ShortName) == strings.ToLower(payload[0]) {
			giveawayID = giveawayrecord.ID
			giveawayownerID = giveawayrecord.OwnerID
		}
	}
	winneroutput := ""
	winnerID, err := h.PickWinner(giveawayID)
	if err == nil {
		winneroutput = " and <@" + winnerID + "> has won, Congratulations! Please contact <@" + giveawayownerID + "> to claim your prize!"

	} else if err.Error() == "no entries" {
		winneroutput = " without a winner because no entries were found."
	}

	found := h.DeactivateGiveaway(strings.ToLower(payload[0]), winnerID)

	if found {
		s.ChannelMessageSend(m.ChannelID, payload[0]+" has been ended manually by "+m.Author.Mention()+winneroutput)
		return
	} else {
		s.ChannelMessageSend(m.ChannelID, "Error: No record with short name "+payload[0]+" found.")
		return
	}
}

func (h *GiveawayHandler) DeactivateGiveaway(shortname string, winnerID string) (found bool) {
	found = false
	giveawayrecords, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		return false
	}

	for _, giveawayrecord := range giveawayrecords {
		if strings.ToLower(giveawayrecord.ShortName) == strings.ToLower(shortname) {
			found = true
			giveawayrecord.Active = false
			giveawayrecord.WinnerID = winnerID
			giveawayrecord.ShortName = "inactive_" + giveawayrecord.ShortName
			h.giveawaydb.UpdateGiveawayRecord(giveawayrecord)
		}
	}

	return found
}

func (h *GiveawayHandler) ListGiveaways(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	records, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading database: "+err.Error())
		return
	}
	output := "Current Active Giveaways:\n```\n"
	for _, record := range records {
		if record.Active {
			output = output + record.ShortName + " - " + record.Description + "\n"
			endTime := record.CreatedDate.Add(record.Duration)
			loc, _ := time.LoadLocation("America/Chicago")
			output = output + "Started: " + record.CreatedDate.In(loc).Format("Mon Jan _2 03:04 MST 2006")
			output = output + "\n"
			output = output + "Ends on: " + endTime.In(loc).Format("Mon Jan _2 03:04 MST 2006")
			output = output + "\n\n"
		}
	}
	output = output + "```"

	s.ChannelMessageSend(m.ChannelID, output)
	return
}

func (h *GiveawayHandler) GiveawayHistory(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'history' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *GiveawayHandler) GiveawayInfo(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'info' expects an argument.")
		return
	}

	giveawayrecords, err := h.giveawaydb.GetAllGiveawayDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not read database: "+err.Error())
		return
	}
	found := false
	output := ":bulb: " + payload[0] + " Giveaway :\n```\n"
	for _, giveawayrecord := range giveawayrecords {
		if strings.ToLower(giveawayrecord.ShortName) == strings.ToLower(payload[0]) {

			entryrecords, err := h.giveawaydb.GetAllEntryDB()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not read database: "+err.Error())
				return
			}
			entrycount := strconv.Itoa(len(entryrecords))
			ownerObject, err := s.User(giveawayrecord.OwnerID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error parsing record: "+err.Error())
				return
			}

			found = true

			//output = output + "Database ID: " + giveawayrecord.ID + "\n"
			output = output + "Name:        " + giveawayrecord.ShortName + "\n"
			output = output + "Owner:       " + ownerObject.Username + "\n"
			output = output + "Entries:     " + entrycount + "\n"
			if !giveawayrecord.Active {
				winnerObject, err := s.User(giveawayrecord.WinnerID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error parsing record: "+err.Error())
					return
				}
				output = output + "Winner:      " + winnerObject.Username + " - (" + winnerObject.ID + ")\n"
			}
			output = output + "Active:      " + strconv.FormatBool(giveawayrecord.Active) + "\n"
			output = output + "Description: " + giveawayrecord.Description + "\n"

			endTime := giveawayrecord.CreatedDate.Add(giveawayrecord.Duration)
			loc, _ := time.LoadLocation("America/Chicago")
			output = output + "Started:     " + giveawayrecord.CreatedDate.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"
			output = output + "Ends on:     " + endTime.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"

		}
	}
	if !found {
		s.ChannelMessageSend(m.ChannelID, "Record for "+payload[0]+" not found!")
		return
	}

	output = output + "```\n"

	s.ChannelMessageSend(m.ChannelID, output)
	return
}
