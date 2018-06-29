package main

import (
	"errors"
	"sync"
	"time"
	"sort"
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
	NewName     string `json:"newname"`
	Managed     bool   `json:"managed"`
	Mentionable bool   `json:"mentionable"`
	Hoist       bool   `json:"hoist"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Timeout     string `json:"Timeout"`
	TimeoutDuration time.Duration
	MemberList  []string
}

type RoleQueued struct {
	ID          string `storm:"id"`
	Remove      bool
	UserID      string
	RoleID      string
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

	RecordList = h.SortRolesList(RecordList)

	return RecordList, nil
}

func (h *RoleDB) SortRolesList(recordlist []RoleRecord) (list []RoleRecord) {

	sort.Slice(recordlist[:], func(i, j int) bool {return recordlist[i].TimeoutDuration > recordlist[j].TimeoutDuration})
	return recordlist
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





// Queued Roles

// AddGiveawayToDB function
func (h *RoleDB) AddRoleQueuedToDB(record RoleQueued) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesQueuedDB")
	err = db.Save(&record)
	return err
}

// RemoveGiveawayFromDB function
func (h *RoleDB) RemoveRoleQueuedFromDB(record RoleQueued) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesQueuedDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDB function
func (h *RoleDB) FlushDB() (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	err = h.db.rawdb.Drop("RolesQueuedDB")
	return err
}

// RemoveGiveawayFromDBByID function
func (h *RoleDB) RemoveRoleQueuedFromDBByID(recordID string) (err error) {

	queuedRole, err := h.GetRoleFromDB(recordID)
	if err != nil {
		return err
	}

	err = h.RemoveRoleRecordFromDB(queuedRole)
	if err != nil {
		return err
	}

	return nil
}

// GetGiveawayFromDB function
func (h *RoleDB) GetRoleQueuedFromDB(recordID string) (Record RoleQueued, err error) {

	RolesQueuedDB, err := h.GetAllRoleQueuedDB()
	if err != nil {
		return Record, err
	}

	for _, i := range RolesQueuedDB {
		if i.ID == recordID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllGiveawayDB function
func (h *RoleDB) GetAllRoleQueuedDB() (RecordList []RoleQueued, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesQueuedDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}

func (h *RoleDB) UpdateRoleQueued(record RoleQueued) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RolesQueuedDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}