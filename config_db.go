package main

import (
	"sync"
	"github.com/bwmarrin/discordgo"
	"errors"
)

// Notifications struct
type ConfigDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

type ConfigEntry struct {
	Name        string `storm:"id"`// The name of our config option
	Setting     string
	Value      int
	Enabled      bool
}

// AddNotificationToDB function
func (h *ConfigDB) AddConfigToDB(config ConfigEntry) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Config")
	err = db.Save(&config)
	return err
}

// RemoveNotificationFromDB function
func (h *ConfigDB) RemoveConfigFromDB(config ConfigEntry) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Config")
	err = db.DeleteStruct(&config)
	return err
}



// RemoveNotificationFromDBByID function
func (h *ConfigDB) RemoveConfigFromDBByName(configname string, s *discordgo.Session) (err error) {

	entry, err := h.GetConfigFromDB(configname)
	if err != nil {
		return err
	}

	err = h.RemoveConfigFromDB(entry)
	if err != nil {
		return err
	}

	return nil
}

// GetNotificationFromDB function
func (h *ConfigDB) GetConfigFromDB(configname string) (entry ConfigEntry, err error) {

	configs, err := h.GetAllConfigs()
	if err != nil {
		return entry, err
	}

	for _, i := range configs {
		if i.Name == configname {
			return i, nil
		}
	}

	return entry, errors.New("No record found")
}

// GetAllNotifications function
func (h *ConfigDB) GetAllConfigs() (configlist []ConfigEntry, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Config")
	err = db.All(&configlist)
	if err != nil {
		return configlist, err
	}

	return configlist, nil
}

// UpdateConfig - Don't use this, it doesn't work as intended
func (h *ConfigDB) UpdateConfig(entry ConfigEntry) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Config")

	err = db.Update(&entry)
	if err != nil {
		return err
	}

	return nil
}


func (h *ConfigDB) CheckEnabled(configname string) (enabled bool, err error) {
	entry, err := h.GetConfigFromDB(configname)
	if err != nil {
		return false, err
	}

	return entry.Enabled, nil
}

func (h *ConfigDB) GetValue(configname string) (value int, err error) {
	entry, err := h.GetConfigFromDB(configname)
	if err != nil {
		return 0, err
	}

	return entry.Value, nil
}