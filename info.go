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
	RecordType  string `json:"recordtype"`
	ImageURL    string `json:"imageurl"`
	Color       int `json:"color"`

	Satellite   SatelliteRecord
	Element     ElementRecord
	Resource    ResourceRecord
	Skill       SkillRecord
}

type SatelliteRecord struct {

	DiscoveredBy string
	SystemZone string
	NotableElements string
	Atmosphere string
	Gravity    string
	SurfaceArea string
	SatelliteCount string
	Satellites []SatelliteRecord
	Biosphere string
	Territories string
	TerritoriesClaimed string
	TerraNullius string

	SatelliteType string `json:"satellitetype"`

	UserList     []string `json:"users"`
	LastWho      time.Time `json:"lastwho"`

}

type ElementRecord struct {

}

type ResourceRecord struct {

}

type SkillRecord struct {

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