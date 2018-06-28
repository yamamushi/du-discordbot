package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

// GiveawayDB struct
type GiveawayDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// GiveawayRecord struct
type GiveawayRecord struct {
	ID      string `storm:"id"`
	OwnerID  string
	ShortName string
	Description string
	CreatedDate    time.Time
	Duration time.Duration
	Active bool
}

// GiveawayEntry struct
type GiveawayEntry struct {
	ID      string `storm:"id"`
	GiveawayID string `storm:"id"`
	UserID  string
	Date    time.Time
}


// AddGiveawayToDB function
func (h *GiveawayDB) AddGiveawayRecordToDB(record GiveawayRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayDB")
	err = db.Save(&record)
	return err
}

// RemoveGiveawayFromDB function
func (h *GiveawayDB) RemoveGiveawayRecordFromDB(record GiveawayRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDBByID function
func (h *GiveawayDB) RemoveGiveawayRecordFromDBByID(giveawayID string, s *discordgo.Session) (err error) {

	Giveaway, err := h.GetGiveawayFromDB(giveawayID)
	if err != nil {
		return err
	}

	err = h.RemoveGiveawayRecordFromDB(Giveaway)
	if err != nil {
		return err
	}

	return nil
}

// GetGiveawayFromDB function
func (h *GiveawayDB) GetGiveawayFromDB(giveawayID string) (Record GiveawayRecord, err error) {

	GiveawayDB, err := h.GetAllGiveawayDB()
	if err != nil {
		return Record, err
	}

	for _, i := range GiveawayDB {
		if i.ID == giveawayID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllGiveawayDB function
func (h *GiveawayDB) GetAllGiveawayDB() (RecordList []GiveawayRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}


func (h *GiveawayDB) UpdateGiveawayRecord(record GiveawayRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}