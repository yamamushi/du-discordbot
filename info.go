package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

type InfoDBInterface struct {
	db *DBHandler
	querylocker sync.Mutex
	conf *Config
}

type InfoRecord struct {
	Name        string `storm:"id",json:"userid"`
	Description string `json:"description"`
	IsLocation          bool `json:"islocation"`
	IsResource         bool `json:"isresource"`
	IsElement         bool `json:"iselement"`
	IsSkill     bool `json:"isskill"`
	ResourceList []string `json:"resourcelist"`
	SkillList    []string `json:"skilllist"`
	UserList     []string `json:"users"`
	LastWho      time.Time `json:"lastwho"`
	ImageURL    string `json:"imageurl"`
	Color       string `json:"color"`
}



// SaveRecordToDB function
func (h *InfoDBInterface) SaveRecordToDB(record InfoRecord, c mgo.Collection) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	_, err = c.UpsertId(record.Name, record)
	return err
}

// NewPlayerRecord function
func (h *InfoDBInterface) NewInfoRecord(name string, c mgo.Collection) (err error) {

	record := InfoRecord{Name: name}
	err = h.SaveRecordToDB(record, c)
	return err

}

// GetRecordFromDB function
func (h *InfoDBInterface) GetRecordFromDB(name string, c mgo.Collection) (record InfoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	inforecord := InfoRecord{}
	err = c.Find(bson.M{"name": name}).One(&inforecord)
	return inforecord, err
}

// BackerInterface function
func (h *InfoDBInterface) GetAllBackers(c mgo.Collection) (records []InfoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	err = c.Find(bson.M{}).All(&records)
	return records, err
}