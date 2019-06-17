package main

import (
	"errors"
	"sync"
)

// GalleryDB struct
type GalleryDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

type GalleryConfig struct {
	ChannelID        string `storm:"id"` // The name of our config option
	Whitelist   []string
	Enabled     bool
}

// AddNotificationToDB function
func (h *GalleryDB) AddConfigToDB(config GalleryConfig) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GalleryConfig")
	err = db.Save(&config)
	return err
}

// RemoveNotificationFromDB function
func (h *GalleryDB) RemoveConfigFromDB(config GalleryConfig) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GalleryConfig")
	err = db.DeleteStruct(&config)
	return err
}

// GetNotificationFromDB function
func (h *GalleryDB) GetConfigFromDB(channelid string) (entry GalleryConfig, err error) {

	configs, err := h.GetAllConfigs()
	if err != nil {
		return entry, err
	}

	for _, i := range configs {
		if i.ChannelID == channelid {
			return i, nil
		}
	}

	return entry, errors.New("No record found")
}


// GetAllNotifications function
func (h *GalleryDB) GetAllConfigs() (configlist []GalleryConfig, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GalleryConfig")
	err = db.All(&configlist)
	if err != nil {
		return configlist, err
	}

	return configlist, nil
}


// UpdateConfig - Don't use this, it doesn't work as intended
func (h *GalleryDB) UpdateConfig(config GalleryConfig) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GalleryConfig")

	err = db.Update(&config)
	if err != nil {
		return err
	}

	return nil
}

func (h *GalleryDB) CheckEnabled(channelid string) (enabled bool, err error) {
	entry, err := h.GetConfigFromDB(channelid)
	if err != nil {
		return false, err
	}

	return entry.Enabled, nil
}

func (h *GalleryDB) GetWhitelist(channelid string) (whitelist []string, err error) {
	entry, err := h.GetConfigFromDB(channelid)
	if err != nil {
		return []string{""}, err
	}

	return entry.Whitelist, nil
}