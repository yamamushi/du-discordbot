package main

import (
	"sync"
	"errors"
)

// StateDB struct
type StateDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// StateRecord struct
type StateRecord struct {
	ID      string `storm:"id"`
	LastRecruitmentIDPosted  string
}

func (h *StateDB) GetState()(state StateRecord, err error) {
	statedb, err := h.GetAllStateDB()
	if len(statedb) < 1 {

		uuid, err := GetUUID()
		if err != nil {
			return state, err
		}
		state := StateRecord{ID:uuid}
		err = h.AddStateRecordToDB(state)
		if err != nil {
			return state, err
		}

		return state, nil
	}
	return statedb[0], nil
}

func (h *StateDB) SetState(state StateRecord)(err error) {
	statedb, err := h.GetAllStateDB()
	if len(statedb) < 1 {
		err = h.AddStateRecordToDB(state)
		if err != nil {
			return err
		}
		return nil
	}

	err = h.RemoveStateRecordFromDB(statedb[0])
	if err != nil {
		return err
	}

	err = h.AddStateRecordToDB(state)
	if err != nil {
		return err
	}

	return nil

}

// AddStateToDB function
func (h *StateDB) AddStateRecordToDB(record StateRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("StateDB")
	err = db.Save(&record)
	return err
}

// RemoveStateFromDB function
func (h *StateDB) RemoveStateRecordFromDB(record StateRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("StateDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveStateFromDBByID function
func (h *StateDB) RemoveStateRecordFromDBByID(giveawayID string) (err error) {

	State, err := h.GetStateFromDB(giveawayID)
	if err != nil {
		return err
	}

	err = h.RemoveStateRecordFromDB(State)
	if err != nil {
		return err
	}

	return nil
}

// GetStateFromDB function
func (h *StateDB) GetStateFromDB(giveawayID string) (Record StateRecord, err error) {

	StateDB, err := h.GetAllStateDB()
	if err != nil {
		return Record, err
	}

	for _, i := range StateDB {
		if i.ID == giveawayID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllStateDB function
func (h *StateDB) GetAllStateDB() (RecordList []StateRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("StateDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}


func (h *StateDB) UpdateStateRecord(record StateRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("StateDB")

	err = db.DeleteStruct(&record)
	if err != nil {
		return err
	}
	err = db.Save(&record)
	return err
}