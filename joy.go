package main

import (
	"errors"
	"sync"
)

// GiveawayDB struct
type JoyDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// GiveawayRecord struct
type JoyRecord struct {
	UserID  string `storm:"id"`
	Enabled bool
}

// AddGiveawayToDB function
func (h *JoyDB) AddJoyRecordToDB(record JoyRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("JoyDB")
	err = db.Save(&record)
	return err
}

// RemoveGiveawayFromDB function
func (h *JoyDB) RemoveJoyRecordFromDB(record JoyRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("JoyDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDBByID function
func (h *JoyDB) RemoveGiveawayRecordFromDBByID(userID string) (err error) {

	joy, err := h.GetJoyIDFromDB(userID)
	if err != nil {
		return err
	}

	err = h.RemoveJoyRecordFromDB(joy)
	if err != nil {
		return err
	}

	return nil
}

// GetGiveawayFromDB function
func (h *JoyDB) GetJoyIDFromDB(userID string) (Record JoyRecord, err error) {

	JoyDB, err := h.GetAllJoyDB()
	if err != nil {
		return Record, err
	}

	for _, i := range JoyDB {
		if i.UserID == userID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllGiveawayDB function
func (h *JoyDB) GetAllJoyDB() (RecordList []JoyRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("JoyDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}

func (h *JoyDB) UpdateJoyRecord(record JoyRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("JoyDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}
