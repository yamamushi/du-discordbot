package main

import (
	"sync"
)

// Notifications struct
type Stats struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

type StatRecord struct {


}

// AddStatToDB function
func (h *Stats) AddStatToDB(stat StatRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Statistics")
	err = db.Save(&stat)
	return err
}


// RemoveStatFromDB function
func (h *Stats) RemoveStatFromDB(stat StatRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Statistics")
	err = db.DeleteStruct(&stat)
	return err
}