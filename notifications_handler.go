package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"
)

// NotificationsHandler struct
type NotificationsHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
}

// Init function
func (h *NotificationsHandler) Init() {
	h.RegisterCommands()

}

// RegisterCommands function
func (h *NotificationsHandler) RegisterCommands() (err error) {

	h.registry.Register("notifications", "Manage notifications for this channel", "enable(for)|disable(for)|add|remove|list|view|channel|messages")
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

	// Channel Specific Commands

	if command[1] == "enablefor" && len(command) <= 3 {
		s.ChannelMessageSend(m.ChannelID, "Command expects three arguments - ChannelID, Notification ID and Time Interval in hours and minutes. Expected format for the time interval is hours(h) and/or minutes(m) separated with a space, ie: 2h 4m")
		return
	}

	if command[1] == "enablefor" && len(command) > 3 {
		h.EnableChannelNotificationFor(command, s, m)
		return
	}

	if command[1] == "disablefor" && len(command) <= 3 {
		s.ChannelMessageSend(m.ChannelID, "Command expects two arguments")
		return
	}

	if command[1] == "disablefor" && len(command) > 3 {
		h.DisableChannelNotificationFor(command[2], command[3], s, m)
		return
	}

	if command[1] == "view" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "view" && len(command) > 2 {
		h.ViewNotificationMessageID(command[2], s, m)
		return
	}

	if command[1] == "messagelist" || command[1] == "listmessages" || command[1] == "listnotifications" || command[1] == "messages" {
		if len(command) > 2 {
			page := command[2]
			h.GetAllNotifications(page, s, m)
			return
		}
		h.GetAllNotifications("1", s, m)
		return
	}

	if command[1] == "list" {
		if len(command) > 2 {
			page := command[2]
			h.GetAllChannelNotifications(page, s, m)
			return
		}
		h.GetAllChannelNotifications("1", s, m)
		return
	}

	if command[1] == "linked" && len(command) <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "linked" && len(command) > 2 {
		h.GetAllLinkedChannels(command[2], s, m)
		return
	}

	if command[1] == "channel" || command[1] == "channellist" || command[1] == "listchannel" {
		if len(command) > 3 {
			page := command[3]
			h.GetAllChannelNotificationsFor(command[2], page, s, m)
			return
		}
		h.GetAllChannelNotificationsFor(command[2], "1", s, m)
		return
	}

	if command[1] == "flush" {
		notificationsdb := Notifications{db: h.db}
		err := notificationsdb.FlushChannelNotifications(m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error Flushing DB: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Channel Notifications Cleared")
		return
	}
}

// AddNotification function
func (h *NotificationsHandler) AddNotification(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	message := ""
	for i, text := range command {

		if i > 1 {
			message = message + text + " "
		}
	}

	notificationsdb := Notifications{db: h.db}
	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	id := strings.Split(uuid, "-")

	notification := Notification{ID: id[0], Message: message}

	err = notificationsdb.AddNotificationToDB(notification)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error adding notification to db: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Added message to notifications list with id: "+id[0])
	return

}

// RemoveNotification fiunction
func (h *NotificationsHandler) RemoveNotification(messageid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	notificationsdb := Notifications{db: h.db}

	err := notificationsdb.RemoveNotificationFromDBByID(messageid, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Message successfully removed from database.")
	return
}

// GetAllNotifications function
func (h *NotificationsHandler) GetAllNotifications(page string, s *discordgo.Session, m *discordgo.MessageCreate) {

	pagenum, err := strconv.Atoi(page)
	if err != nil {
		pagenum = 1
	}

	notificationsdb := Notifications{db: h.db}

	notificationlist, err := notificationsdb.GetAllNotifications()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving notifications list: "+err.Error())
		return
	}

	pages := (len(notificationlist) / 5)
	if len(notificationlist)%5 != 0 {
		pages = pages + 1
	}

	if pagenum > pages {
		pagenum = pages
	}

	list := "```"
	list = list + "   ID    | Message\n"
	list = list + "---------------------------------------------------------------------------\n"
	count := 0
	for num, notification := range notificationlist {
		if num >= ((pagenum * 5) - 5) {

			count = count + 1

			output := notification.ID + " | " + notification.Message + "\n"
			list = list + output
			list = list + "---------------------------------------------------------------------------\n"

			if count == 5 {
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

// GetAllChannelNotifications function
func (h *NotificationsHandler) GetAllChannelNotifications(page string, s *discordgo.Session, m *discordgo.MessageCreate) {

	pagenum, err := strconv.Atoi(page)
	if err != nil {
		pagenum = 1
	}

	notificationsdb := Notifications{db: h.db}

	notificationlist, err := notificationsdb.GetAllChannelNotifications()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving notifications list: "+err.Error())
		return
	}

	listsize := 0
	for _, notification := range notificationlist {
		if notification.ChannelID == m.ChannelID {
			listsize = listsize + 1
		}
	}

	pages := (listsize / 5)
	if listsize%5 != 0 {
		pages = pages + 1
	}

	if pagenum > pages {
		pagenum = pages
	}

	list := "```\n"
	list = list + "   ID    | Interval | MessageID |      Last Run       |      Next Run\n"
	list = list + "---------------------------------------------------------------------------\n"
	count := 0
	for num, notification := range notificationlist {

		if notification.ChannelID == m.ChannelID {

			if num >= ((pagenum * 5) - 5) {

				/*
					notificationmessage, err := notificationsdb.GetNotificationFromDB(notification.Notification)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Error retrieving notification from db")
						return
					}*/
				count = count + 1

				timeout := notification.Timeout
				hours := "0h"
				minutes := "0m"
				holder := strings.Split(timeout, " ")

				for i := 0; i < len(holder); i++ {
					if strings.Contains(holder[i], "h") {
						hours = holder[i]
					} else if strings.Contains(holder[i], "m") {
						minutes = holder[i]
					}
				}

				timeout = hours + " " + minutes

				if len(timeout) < 10 {
					padright := 9 - len(timeout)
					for i := 0; i < padright; i++ {
						timeout = timeout + " "
					}
				}

				lastrun := notification.LastRun.UTC()
				lastrunstring := lastrun.Format("2006-01-02 15:04:05")

				hoursint, minutesint, err := h.ParseTimeout(timeout)
				if err != nil {
					fmt.Println("Error parsing timeout for channel notification: " + notification.ID)
				}

				if hoursint > 0 {
					minutesint = (hoursint * 60) + minutesint
				}

				interval := time.Duration(minutesint * 60 * 1000 * 1000 * 1000)
				nextruntime := notification.LastRun.Add(interval)
				nextrun := nextruntime.UTC().Format("2006-01-02 15:04:05")

				notificationid := notification.Notification

				/*
					message := notificationmessage.Message
					if len(message) > 13 {
							message = truncateString(message, 13)
					}*/

				output := notification.ID + " | " + timeout + "| " + notificationid + "  | " + lastrunstring + " | " + nextrun + "\n"
				list = list + output
				list = list + "---------------------------------------------------------------------------\n"

				if count == 5 {
					list = list + "```"
					list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
					s.ChannelMessageSend(m.ChannelID, list)
					return
				}
			}
		}
	}

	list = list + "```"
	list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
	s.ChannelMessageSend(m.ChannelID, list)
	return
}

// EnableChannelNotification function
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
	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	id := strings.Split(uuid, "-")
	err = notificationsdb.CreateChannelNotification(id[0], command[2], m.ChannelID, parsed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Notification enabled for this channel.")
	return
}

// DisableChannelNotification function
func (h *NotificationsHandler) DisableChannelNotification(notificationid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	notificationsdb := Notifications{db: h.db}
	err := notificationsdb.RemoveChannelNotificationFromDBByID(notificationid, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Notification disabled for this channel.")
	return
}

// Channel Specific Functions

// GetAllChannelNotifications function
func (h *NotificationsHandler) GetAllChannelNotificationsFor(channelname string, page string, s *discordgo.Session, m *discordgo.MessageCreate) {

	channelname = strings.TrimPrefix(channelname, "<#")
	channelname = strings.TrimSuffix(channelname, ">")

	channel, err := s.Channel(channelname)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not parse channel (use the # name not the id number):"+err.Error())
		return
	}

	channelmention, err := MentionChannel(channel.ID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing channel: "+err.Error())
		return
	}

	pagenum, err := strconv.Atoi(page)
	if err != nil {
		pagenum = 1
	}

	notificationsdb := Notifications{db: h.db}

	notificationlist, err := notificationsdb.GetAllChannelNotifications()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving notifications list: "+err.Error())
		return
	}

	listsize := 0
	for _, notification := range notificationlist {
		if notification.ChannelID == channel.ID {
			listsize = listsize + 1
		}
	}

	pages := (listsize / 5)
	if listsize%5 != 0 {
		pages = pages + 1
	}

	if pagenum > pages {
		pagenum = pages
	}

	list := ":bulb: Channel Notifications for " + channelmention + ": \n"
	list = list + "```\n"
	list = list + "   ID    | Interval | MessageID |      Last Run       |      Next Run\n"
	list = list + "---------------------------------------------------------------------------\n"
	count := 0
	for num, notification := range notificationlist {

		if notification.ChannelID == channel.ID {

			if num >= ((pagenum * 5) - 5) {
				count = count + 1

				/*
					notificationmessage, err := notificationsdb.GetNotificationFromDB(notification.Notification)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Error retrieving notification from db")
						return
					}*/

				timeout := notification.Timeout
				hours := "0h"
				minutes := "0m"
				holder := strings.Split(timeout, " ")

				for i := 0; i < len(holder); i++ {
					if strings.Contains(holder[i], "h") {
						hours = holder[i]
					} else if strings.Contains(holder[i], "m") {
						minutes = holder[i]
					}
				}

				timeout = hours + " " + minutes

				if len(timeout) < 10 {
					padright := 9 - len(timeout)
					for i := 0; i < padright; i++ {
						timeout = timeout + " "
					}
				}

				lastrun := notification.LastRun.UTC()
				lastrunstring := lastrun.Format("2006-01-02 15:04:05")

				hoursint, minutesint, err := h.ParseTimeout(timeout)
				if err != nil {
					fmt.Println("Error parsing timeout for channel notification: " + notification.ID)
				}

				if hoursint > 0 {
					minutesint = (hoursint * 60) + minutesint
				}

				interval := time.Duration(minutesint * 60 * 1000 * 1000 * 1000)
				nextruntime := notification.LastRun.Add(interval)
				nextrun := nextruntime.UTC().Format("2006-01-02 15:04:05")

				notificationid := notification.Notification

				/*
					message := notificationmessage.Message
					if len(message) > 13 {
						message = truncateString(message, 13)
					}*/

				output := notification.ID + " | " + timeout + "| " + notificationid + "  | " + lastrunstring + " | " + nextrun + "\n"
				list = list + output
				list = list + "---------------------------------------------------------------------------\n"

				if count == 5 {
					list = list + "```"
					list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
					s.ChannelMessageSend(m.ChannelID, list)
					return
				}
			}
		}
	}

	list = list + "```"
	list = list + "Page " + strconv.Itoa(pagenum) + " of " + strconv.Itoa(pages)
	s.ChannelMessageSend(m.ChannelID, list)
	return
}

// EnableChannelNotificationFor function
func (h *NotificationsHandler) EnableChannelNotificationFor(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	var parsed string

	for i, field := range command {
		if i > 3 {
			parsed = parsed + field + " "
		}
	}

	channelname := command[2]
	channelname = strings.TrimPrefix(channelname, "<#")
	channelid := strings.TrimSuffix(channelname, ">")

	_, err := s.Channel(channelid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not parse channel (use the # name not the id number):"+err.Error())
		return
	}

	channelmention, err := MentionChannel(channelid, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing channel: "+err.Error())
		return
	}

	_, _, err = h.ParseTimeout(parsed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	notificationsdb := Notifications{db: h.db}
	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	id := strings.Split(uuid, "-")
	err = notificationsdb.CreateChannelNotification(id[0], command[3], channelid, parsed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Notification enabled for "+channelmention)
	return
}

// DisableChannelNotificationFor function
func (h *NotificationsHandler) DisableChannelNotificationFor(channelname string, notificationid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	channelname = strings.TrimPrefix(channelname, "<#")
	channelid := strings.TrimSuffix(channelname, ">")

	_, err := s.Channel(channelid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not parse channel (use the # name not the id number):"+err.Error())
		return
	}

	channelmention, err := MentionChannel(channelid, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing channel: "+err.Error())
		return
	}

	notificationsdb := Notifications{db: h.db}
	err = notificationsdb.RemoveChannelNotificationFromDBByID(notificationid, channelid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Notification disabled for "+channelmention)
	return
}

// ViewNotificationMessageID function
func (h *NotificationsHandler) ViewNotificationMessageID(notificationid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	notificationsdb := Notifications{db: h.db}
	notification, err := notificationsdb.GetNotificationFromDB(notificationid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	output := "```"
	output = output + "   ID    | Message\n"
	output = output + "---------------------------------------------------------------------------\n"
	output = output + notification.ID + " | " + notification.Message + "\n"
	output = output + "```"

	s.ChannelMessageSend(m.ChannelID, output)
	return
}

// ViewNotificationMessageID function
func (h *NotificationsHandler) GetAllLinkedChannels(notificationid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	notificationsdb := Notifications{db: h.db}
	notification, err := notificationsdb.GetNotificationFromDB(notificationid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	linkedchannels, err := notificationsdb.GetNotificationLinkedChannels(notification.ID, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	output := "Linked channels for notification id: " + notificationid + " - " + linkedchannels + "\n"

	s.ChannelMessageSend(m.ChannelID, output)
	return
}

// CheckNotifications function
func (h *NotificationsHandler) CheckNotifications(s *discordgo.Session) {

	for true {
		// Only run every X minutes
		time.Sleep(h.conf.DUBotConfig.Notifications * time.Minute)

		//fmt.Println("Running Notifications Handler")
		notificationsdb := Notifications{db: h.db}

		notificationlist, err := notificationsdb.GetAllChannelNotifications()
		if err != nil {
			fmt.Println("Error reading from channel notifications database: " + err.Error())
		}

		for _, channelnotification := range notificationlist {

			//fmt.Println("Reading Notification")

			timeout := channelnotification.Timeout

			hours, minutes, err := h.ParseTimeout(timeout)
			if err != nil {
				fmt.Println("Error parsing timeout for channel notification: " + channelnotification.ID)
			}

			if hours > 0 {
				minutes = (hours * 60) + minutes
			}

			interval := time.Duration(minutes * 60 * 1000 * 1000 * 1000)
			//fmt.Println("Interval: " + strconv.Itoa(interval))

			diff := time.Now().Sub(channelnotification.LastRun)
			if diff > interval {

				notification, err := notificationsdb.GetNotificationFromDB(channelnotification.Notification)
				if err != nil {
					fmt.Println("Error reading from notifications database: " + err.Error())
				}
				s.ChannelMessageSend(channelnotification.ChannelID, notification.Message)

				channelnotification.LastRun = time.Now()
				notificationsdb.UpdateChannelNotification(channelnotification)
			}
		}
	}

}

// ParseTimeout function
func (h *NotificationsHandler) ParseTimeout(timeout string) (hours int64, minutes int64, err error) {

	hoursstring := "0"
	minutesstring := "0"

	if !strings.Contains(timeout, " ") {
		return 0, 0, errors.New("Invalid time interval format")
	}

	separated := strings.Split(timeout, " ")

	for _, field := range separated {

		for _, value := range field {
			switch {
			case value >= '0' && value <= '9':
				if strings.Contains(field, "h") {
					hoursstring = strings.TrimSuffix(field, "h")
					hours, err = strconv.ParseInt(hoursstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse hours")
					}
				} else if strings.Contains(field, "m") {
					minutesstring = strings.TrimSuffix(field, "m")
					minutes, err = strconv.ParseInt(minutesstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse minutes")
					}
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

	if hours == 0 && minutes == 0 {
		return hours, minutes, errors.New("Invalid interval specified")
	}

	return hours, minutes, nil
}
