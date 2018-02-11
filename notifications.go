package main

import (
	"sync"
	"time"
	"errors"
	"github.com/bwmarrin/discordgo"
)

// Notifications struct
type Notifications struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// RSSFeed struct
type Notification struct {
	ID          string `storm:"id"`
	Message     string
}

type ChannelNotification struct {

	ID          	string `storm:"id"`
	Notification	string	// The ID of our notification message
	ChannelID   	string `storm:"index"` // Limit our notifications per channel
	LastRun    		time.Time
	Timeout         string    `storm:"index"`

}


// AddNotificationToDB function
func (h *Notifications) AddNotificationToDB(notification Notification) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Notifications")
	err = db.Save(&notification)
	return err
}


// RemoveNotificationFromDB function
func (h *Notifications) RemoveNotificationFromDB(notification Notification) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Notifications")
	err = db.DeleteStruct(&notification)
	return err
}

// RemoveNotificationFromDBByID function
func (h *Notifications) RemoveNotificationFromDBByID(messageid string, s *discordgo.Session) (err error) {

	notification, err := h.GetNotificationFromDB(messageid)
	if err != nil {
		return err
	}

	channelnotifications, err := h.GetAllChannelNotifications()
	if err != nil {
		return err
	}

	found := false
	channelsinuse := ""
	for _, channelnotification := range channelnotifications {

		if channelnotification.Notification == notification.ID {
			found = true
			mention, err := MentionChannel(channelnotification.ChannelID, s)
			if err != nil {
				return err
			}

			if channelsinuse != ""{
				channelsinuse = channelsinuse + ", " + mention
			} else {
				channelsinuse = mention
			}
		}
	}
	if (found) {
		return errors.New("Could not remove message from db, it is currently in use by the following channel(s):\n" + channelsinuse)
	}

	err = h.RemoveNotificationFromDB(notification)
	if err != nil {
		return err
	}

	return nil
}

// GetNotificationFromDB function
func (h *Notifications) GetNotificationFromDB(messageid string) (notification Notification, err error) {

	notifications, err := h.GetAllNotifications()
	if err != nil{
		return notification, err
	}

	for _, i := range notifications {
		if i.ID == messageid{
			return i, nil
		}
	}

	return notification, errors.New("No record found")
}


// GetAllNotifications function
func (h *Notifications) GetAllNotifications() (notificationlist []Notification, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Notifications")
	err = db.All(&notificationlist)
	if err != nil {
		return notificationlist, err
	}

	return notificationlist, nil
}





// AddChannelNotificationToDB function
func (h *Notifications) AddChannelNotificationToDB(channelnotification ChannelNotification) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("ChannelNotifications")
	err = db.Save(&channelnotification)
	return err
}

// RemoveChannelNotificationFromDB function
func (h *Notifications) RemoveChannelNotificationFromDB(channelnotification ChannelNotification) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("ChannelNotifications")
	err = db.DeleteStruct(&channelnotification)
	return err
}

func (h *Notifications) UpdateChannelNotification(channelnotification ChannelNotification) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("ChannelNotifications")

	err = db.Update(&channelnotification)
	if err != nil {
		return err
	}

	return nil
}



// RemoveNotificationFromDBByID function
func (h *Notifications) RemoveChannelNotificationFromDBByID(channelnotificationid string, channelid string) (err error) {

	channelnotification, err := h.GetChannelNotificationFromDB(channelnotificationid, channelid)
	if err != nil {
		return err
	}

	err = h.RemoveChannelNotificationFromDB(channelnotification)
	if err != nil {
		return err
	}

	return nil
}


// GetNotificationFromDB function
func (h *Notifications) GetChannelNotificationFromDB(channelnotificationid string, channelid string) (channelnotification ChannelNotification, err error) {

	channelnotifications, err := h.GetAllChannelNotifications()
	if err != nil{
		return channelnotification, err
	}

	for _, i := range channelnotifications {
		if i.ID == channelnotificationid{
			if i.ChannelID == channelid {
				return i, nil
			}
		}
	}

	return channelnotification, errors.New("Channel notification " + channelnotificationid + " does not exist for this channel!")
}


// GetAllChannelNotifications function
func (h *Notifications) GetAllChannelNotifications() (channelnotificationlist []ChannelNotification, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("ChannelNotifications")
	err = db.All(&channelnotificationlist)
	if err != nil {
		return channelnotificationlist, err
	}

	return channelnotificationlist, nil
}


func (h *Notifications) FlushChannelNotifications(channelid string) (err error){

	channelnotifications, err := h.GetAllChannelNotifications()
	if err != nil {
		return err
	}

	for _, notification := range channelnotifications {
		err = h.RemoveChannelNotificationFromDB(notification)
		if err != nil {
			return err
		}
	}

	return nil
}


func (h *Notifications) CreateChannelNotification(id string, notificationid string, channelid string, timeout string) (err error){

	_, err = h.GetNotificationFromDB(notificationid)
	if err != nil {
		return errors.New("Notification ID: " + notificationid + " - not found")
	}

	channelnotifications, err := h.GetAllChannelNotifications()
	if err != nil{
		return err
	}

	for _, i := range channelnotifications {
		idmatches := false
		channelmatches := false
		if i.Notification == notificationid {
			idmatches = true
		}
		if i.ChannelID == channelid {
			channelmatches = true
		}

		if idmatches && channelmatches {
			return errors.New("Notification is already enabled for this channel")
		}
	}

	channelnotification := ChannelNotification{ID: id, Notification: notificationid, ChannelID: channelid, LastRun: time.Now(), Timeout: timeout}

	err = h.AddChannelNotificationToDB(channelnotification)
	if err != nil{
		return err
	}
	return nil
}

