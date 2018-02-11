package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
	"strconv"
	"fmt"
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
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'notifications' command")
		return
	}

	if command[1] == "add" && len(command) == 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "add" && len(command) > 2 {
		h.AddNotification(command, s, m)
		return
	}

	if command[1] == "remove" && len(command) == 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "remove" && len(command) > 2 {
		h.RemoveNotification(command[2], s, m)
		return
	}

	if command[1] == "enable" && len(command) == 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "disable" && len(command) == 2 {
		s.ChannelMessageSend(m.ChannelID, "Command expects an argument")
		return
	}

	if command[1] == "list" {
		if len(command) > 2 {
			page := command[2]
			h.GetAllNotifications(page, s, m)
			return
		}
		h.GetAllNotifications("1", s, m )
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



// CheckNotifications function
func (h *NotificationsHandler) CheckNotifications(s *discordgo.Session) {

	for true {
		// Only run every X minutes
		time.Sleep(h.conf.DUBotConfig.Notifications * time.Minute)

	}

}