package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
	"strconv"
	"fmt"
	"errors"
)

// NotificationsHandler struct
type NotificationsHandler struct {

	conf       *Config
	registry   *CommandRegistry
	callback   *CallbackHandler
	db         *DBHandler

}

// Init function
func (h *NotificationsHandler) Init() {
	h.RegisterCommands()

}

// RegisterCommands function
func (h *NotificationsHandler) RegisterCommands() (err error) {

	h.registry.Register("notifications", "Manage notifications for this channel", "enable|disable|add|remove|list")
	return nil

}


// Read function
func (h *NotificationsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"notifications") {
		if h.registry.CheckPermission("notifications", m.ChannelID, user) {

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
func (h *NotificationsHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(command) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'notifications' command, see usage for more info")
		return
	}

	if command[1] == "add" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "add" && len(command) > 2 {
		h.AddNotification(command, s, m)
		return
	}

	if command[1] == "remove" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "remove" && len(command) > 2 {
		h.RemoveNotification(command[2], s, m)
		return
	}

	if command[1] == "enable" && len(command) <= 3 {
		s.ChannelMessageSend(m.ChannelID, "Command expects two arguments - Notification ID and Time Interval in hours and minutes. Expected format for the time interval is hours(h) and/or minutes(m) separated with a space, ie: 2h 4m")
		return
	}

	if command[1] == "enable" && len(command) > 3 {
		h.EnableChannelNotification(command, s, m)
		return
	}

	if command[1] == "disable" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "disable" && len(command) > 2 {
		h.DisableChannelNotification(command[2], s, m)
		return
	}

	if command[1] == "messagelist" || command[1] == "listmessages" || command[1] == "listnotifications" || command[1] == "messages"{
		if len(command) > 2 {
			page := command[2]
			h.GetAllNotifications(page, s, m)
			return
		}
		h.GetAllNotifications("1", s, m )
		return
	}

	if command[1] == "list" {
		if len(command) > 2 {
			page := command[2]
			h.GetAllChannelNotifications(page, s, m)
			return
		}
		h.GetAllChannelNotifications("1", s, m )
		return
	}
	if command[1] == "flush" {
		notificationsdb := Notifications{db: h.db}
		err := notificationsdb.FlushChannelNotifications(m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error Flushing DB: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Channel Notifications Cleared")
		return
	}
}

// AddNotification function
func (h *NotificationsHandler) AddNotification(command []string, s *discordgo.Session, m *discordgo.MessageCreate){

	message := ""
	for i, text := range command {

		if i > 1 {
			message = message + text + " "
		}
	}

	notificationsdb := Notifications{db: h.db}

	id := strings.Split(GetUUID(), "-")

	notification := Notification{ID: id[0], Message: message}

	err := notificationsdb.AddNotificationToDB(notification)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error adding notification to db: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Added message to notifications list")
	return

}


// RemoveNotification fiunction
func (h *NotificationsHandler) RemoveNotification(messageid string, s *discordgo.Session, m *discordgo.MessageCreate){

	notificationsdb := Notifications{db: h.db}

	err := notificationsdb.RemoveNotificationFromDBByID(messageid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not remove notification from db: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Message successfully removed from database.")
	return
}


// GetAllNotifications function
func (h *NotificationsHandler) GetAllNotifications(page string, s *discordgo.Session, m *discordgo.MessageCreate) {

	fmt.Println("page: " + page)
	pagenum, err := strconv.Atoi(page)
	if err != nil {
		pagenum = 1
	}

	notificationsdb := Notifications{db: h.db}

	notificationlist, err := notificationsdb.GetAllNotifications()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving notifications list: " + err.Error())
		return
	}

	pages := (len(notificationlist) / 10)+1
	if len(notificationlist) == 10 {
		pages = 1
	}

	if pagenum > pages{
		pagenum = 1
	}

	list := "```"
	count := 0
	for num, notification := range notificationlist {

		count = count + 1

		if num >= ((pagenum * 10)-10) {
			output := notification.ID + ": " + notification.Message + "\n"
			list = list + output

			if count == 10{
				list = list + "```"
				list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
				s.ChannelMessageSend(m.ChannelID, list)
				return
			}
		}
	}

	list = list + "```"
	list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
	s.ChannelMessageSend(m.ChannelID, list)
	return
}



// GetAllNotifications function
func (h *NotificationsHandler) GetAllChannelNotifications(page string, s *discordgo.Session, m *discordgo.MessageCreate) {

	fmt.Println("page: " + page)
	pagenum, err := strconv.Atoi(page)
	if err != nil {
		pagenum = 1
	}

	notificationsdb := Notifications{db: h.db}

	notificationlist, err := notificationsdb.GetAllChannelNotifications()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving notifications list: " + err.Error())
		return
	}

	pages := (len(notificationlist) / 10)+1
	if len(notificationlist) == 10 {
		pages = 1
	}

	if pagenum > pages{
		pagenum = 1
	}

	list := "```\n"
	list = list + "   ID    |  Interval  | Message\n"
	list = list + "--------------------------------------------------------------\n"
	count := 0
	for num, notification := range notificationlist {

		count = count + 1

		if num >= ((pagenum * 10)-10) {

			notificationmessage, err := notificationsdb.GetNotificationFromDB(notification.Notification)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving notification from db")
				return
			}

			timeout := notification.Timeout
			hours := "0h"
			minutes := "0m"
			holder := strings.Split(timeout, " ")

			for i := 0; i < len(holder); i++ {
				if strings.Contains(holder[i], "h"){
					hours = holder[i]
				} else if strings.Contains(holder[i], "m"){
					minutes = holder[i]
				}
			}

			timeout = hours + " " + minutes


			if len(timeout) < 12 {
				padright := 11-len(timeout)
				for i := 0; i < padright; i++ {
					timeout = timeout + " "
				}
			}

			output := notification.ID + " | " + timeout + "| " + notificationmessage.Message + "\n"
			list = list + output

			if count == 10{
				list = list + "```"
				list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
				s.ChannelMessageSend(m.ChannelID, list)
				return
			}
		}
	}

	list = list + "```"
	list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
	s.ChannelMessageSend(m.ChannelID, list)
	return
}


func (h *NotificationsHandler) EnableChannelNotification(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {


	var parsed string

	for i, field := range command {
		if i > 2 {
			parsed = parsed + field + " "
		}
	}

	_, _, err := h.ParseTimeout(parsed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	notificationsdb := Notifications{db: h.db}

	id := strings.Split(GetUUID(), "-")
	err = notificationsdb.CreateChannelNotification(id[0], command[2], m.ChannelID, parsed)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Notification enabled for this channel.")
	return
}


func (h *NotificationsHandler) DisableChannelNotification(notificationid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	notificationsdb := Notifications{db: h.db}
	err := notificationsdb.RemoveChannelNotificationFromDBByID(notificationid)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Notification disabled for this channel.")
	return
}


// CheckNotifications function
func (h *NotificationsHandler) CheckNotifications(s *discordgo.Session) {

	for true {
		// Only run every X minutes
		time.Sleep(h.conf.DUBotConfig.Notifications * time.Minute)

	}

}





func (h *NotificationsHandler) ParseTimeout(timeout string) (hours int64, minutes int64, err error){

	hoursstring 	:= "0"
	minutesstring 	:= "0"

	if !strings.Contains(timeout, " "){
		return 0, 0, errors.New("Invalid time interval format")
	}

	separated := strings.Split(timeout, " ")

	for _, field := range separated {

		for _, value := range field {
			switch {
			case value >= '0' && value <= '9':
				if strings.Contains(field, "h"){
					hoursstring = strings.TrimSuffix(field, "h")
					hours, err = strconv.ParseInt(hoursstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse hours")
					}
					fmt.Println("hours:" + minutesstring)

				} else if strings.Contains(field, "m"){
					minutesstring = strings.TrimSuffix(field, "m")
					minutes, err = strconv.ParseInt(minutesstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse minutes")
					}
					fmt.Println("minutes:" + minutesstring)
				} else {
					return 0, 0, errors.New("Invalid time interval format")
				}
				break
			default:
				return 0, 0, errors.New("Invalid time interval format")
			}
			break
		}
	}

	return hours, minutes, nil
}