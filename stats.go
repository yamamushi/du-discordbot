package main

import (
	"sync"
)

// Notifications struct
type StatsDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

type StatRecord struct {
	Date string `json:"date" storm:"id"`
	TotalUsers int `json:"totalusers"`
	OnlineUsers int `json:"onlineusers"`
	IdleUsers int `json:"idleusers"`
	DNDUsers int `json:"dndusers"`
	InvisibleUsers int `json:"invisibleusers"`
	GamingUsers int `json:"gamingusers"`
	VoiceUsers int `json:"voiceusers"`
	MessageCount int `json:"messagecount"`
	Engagement int `json:"engagement"`
	EngagementDaily int `json:"engagementdaily"`
	ActiveUserCount int `json:"activeusercount"`
}

// AddStatToDB function
func (h *StatsDB) AddStatToDB(stat StatRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Statistics")
	err = db.Save(&stat)
	return err
}

// UpdateStatInDB function
func (h *StatsDB) UpdateStatInDB(stat StatRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Statistics")
	err = db.Update(&stat)
	return err
}

// RemoveStatFromDB function
func (h *StatsDB) RemoveStatFromDB(stat StatRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Statistics")
	err = db.DeleteStruct(&stat)
	return err
}

func (h *StatsDB) GetFullDB() (stats []StatRecord, err error){
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Statistics")
	err = db.All(&stats)
	if err != nil {
		return stats, err
	}
	return stats, nil
}
