package main

import (
	"errors"
	"sync"
	"time"
)

// GiveawayDB struct
type RecruitmentDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// GiveawayRecord struct
type RecruitmentRecord struct {
	ID                     string `storm:"id"`
	OwnerID                string
	OrgName                string
	Description            string
	LastRun                time.Time
	LastValidated          time.Time
	ValidationReminderSent bool
	Created                time.Time
	ApprovedDate           time.Time
	Approved               bool
	Approver               string
}

// We store our records in the list so that bot reboots don't break displaying them
type RecruitmentDisplayRecord struct {
	ID            string `storm:"id"`
	RecruitmentID string
}

// AddRecruitmentRecordToDB function
func (h *RecruitmentDB) AddRecruitmentRecordToDB(record RecruitmentRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDB")
	err = db.Save(&record)
	return err
}

// RemoveRecruitmentRecordFromDB function
func (h *RecruitmentDB) RemoveRecruitmentRecordFromDB(record RecruitmentRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDBByID function
func (h *RecruitmentDB) RemoveRecruitmentRecordFromDBByID(recruitmentRecordID string) (err error) {

	recruitmentrecord, err := h.GetRecruitmentRecordFromDB(recruitmentRecordID)
	if err != nil {
		return err
	}

	err = h.RemoveRecruitmentRecordFromDB(recruitmentrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetRecruitmentRecordFromDB function
func (h *RecruitmentDB) GetRecruitmentRecordFromDB(recruitmentRecordID string) (Record RecruitmentRecord, err error) {

	recruitmentdb, err := h.GetAllRecruitmentDB()
	if err != nil {
		return Record, err
	}

	for _, i := range recruitmentdb {
		if i.ID == recruitmentRecordID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllRecruitmentDB function
func (h *RecruitmentDB) GetAllRecruitmentDB() (RecordList []RecruitmentRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}

func (h *RecruitmentDB) UpdateRecruitmentRecord(record RecruitmentRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}

// Display Records

// AddRecruitmentRecordToDB function
func (h *RecruitmentDB) AddRecruitmentDisplayRecordToDB(record RecruitmentDisplayRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDisplayDB")
	err = db.Save(&record)
	return err
}

// RemoveRecruitmentRecordFromDB function
func (h *RecruitmentDB) RemoveRecruitmentDisplayRecordFromDB(record RecruitmentDisplayRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDisplayDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveGiveawayFromDBByID function
func (h *RecruitmentDB) RemoveRecruitmentDisplayRecordFromDBByID(recruitmentDisplayRecordID string) (err error) {

	displayrecord, err := h.GetRecruitmentDisplayRecordFromDB(recruitmentDisplayRecordID)
	if err != nil {
		return err
	}

	err = h.RemoveRecruitmentDisplayRecordFromDB(displayrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetRecruitmentRecordFromDB function
func (h *RecruitmentDB) GetRecruitmentDisplayRecordFromDB(displayrecordID string) (Record RecruitmentDisplayRecord, err error) {

	recruitmentdisplaydb, err := h.GetAllRecruitmentDisplayDB()
	if err != nil {
		return Record, err
	}

	for _, i := range recruitmentdisplaydb {
		if i.ID == displayrecordID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllRecruitmentDB function
func (h *RecruitmentDB) GetAllRecruitmentDisplayDB() (RecordList []RecruitmentDisplayRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDisplayDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}

func (h *RecruitmentDB) UpdateRecruitmentDisplayRecord(record RecruitmentDisplayRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RecruitmentDisplayDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}
