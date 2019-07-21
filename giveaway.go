package main

import (
	"errors"
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
	WinnerID string
	CreatedDate    time.Time
	Duration time.Duration
	Active bool
	Restricted bool
}

// GiveawayEntry struct
type GiveawayEntry struct {
	ID      string `storm:"id"`
	GiveawayID string
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
func (h *GiveawayDB) RemoveGiveawayRecordFromDBByID(giveawayID string) (err error) {

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


// Entry Functions

// AddGiveawayToDB function
func (h *GiveawayDB) AddEntryRecordToDB(record GiveawayEntry) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayEntriesDB")
	err = db.Save(&record)
	return err
}

// RemoveGiveawayFromDB function
func (h *GiveawayDB) RemoveEntryRecordFromDB(record GiveawayEntry) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayEntriesDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDBByID function
func (h *GiveawayDB) RemoveEntryRecordFromDBByID(giveawayEntryID string) (err error) {

	entry, err := h.GetEntryFromDB(giveawayEntryID)
	if err != nil {
		return err
	}

	err = h.RemoveEntryRecordFromDB(entry)
	if err != nil {
		return err
	}

	return nil
}

// GetGiveawayFromDB function
func (h *GiveawayDB) GetEntryFromDB(entryID string) (Record GiveawayEntry, err error) {

	EntryDB, err := h.GetAllEntryDB()
	if err != nil {
		return Record, err
	}

	for _, i := range EntryDB {
		if i.ID == entryID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllGiveawayDB function
func (h *GiveawayDB) GetAllEntryDB() (RecordList []GiveawayEntry, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayEntriesDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}


func (h *GiveawayDB) UpdateEntryRecord(record GiveawayEntry) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("GiveawayEntriesDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}


func (h *GiveawayDB) FlushEntriesForGiveaway(giveawayID string) (err error) {
	entries, err := h.GetAllEntryDB()
	if err != nil {
		return err
	}
	for _, entry := range entries {

		if entry.GiveawayID == giveawayID {
			err = h.RemoveEntryRecordFromDBByID(entry.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *GiveawayDB) EntryCountForGiveaway(giveawayID string) (count int, err error) {
	entries, err := h.GetAllEntryDB()
	if err != nil {
		return 0, err
	}
	count = 0
	for _, entry := range entries {
		if entry.GiveawayID == giveawayID {
			count = count + 1
		}
	}
	return count, nil
}