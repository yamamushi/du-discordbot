package main

import (
	"errors"
	"sync"
	"time"
)

// RoleDB struct
type RoleDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// RoleRecord struct
type RoleRecord struct {
	ID          string `storm:"id" json:"id"`
	Name        string `json:"name"`
	Managed     bool   `json:"managed"`
	Mentionable bool   `json:"mentionable"`
	Hoist       bool   `json:"hoist"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
	Timeout     time.Duration
}


// AddGiveawayToDB function
func (h *RoleDB) AddRoleRecordToDB(record RoleRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesDB")
	err = db.Save(&record)
	return err
}

// RemoveGiveawayFromDB function
func (h *RoleDB) RemoveRoleRecordFromDB(record RoleRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDBByID function
func (h *RoleDB) RemoveRoleRecordFromDBByID(roleID string) (err error) {

	Role, err := h.GetRoleFromDB(roleID)
	if err != nil {
		return err
	}

	err = h.RemoveRoleRecordFromDB(Role)
	if err != nil {
		return err
	}

	return nil
}

// GetGiveawayFromDB function
func (h *RoleDB) GetRoleFromDB(roleID string) (Record RoleRecord, err error) {

	RolesDB, err := h.GetAllRolesDB()
	if err != nil {
		return Record, err
	}

	for _, i := range RolesDB {
		if i.ID == roleID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllGiveawayDB function
func (h *RoleDB) GetAllRolesDB() (RecordList []RoleRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}


func (h *RoleDB) UpdateRoleRecord(record RoleRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}