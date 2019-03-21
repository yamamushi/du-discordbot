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
	RecordType  string `json:"recordtype"`  // satellite/element/resource/skill/user/location
	ThumbnailURL string `json:"thumbnailurl"`
	ImageURL    string `json:"imageurl"`
	Color       int `json:"color"`

	Satellite   SatelliteRecord
	Element     ElementRecord
	Resource    ResourceRecord
	Skill       SkillRecord
	User        UserRecord
	Location    LocationRecord

	EditorID    string `json:"editorid"`
}

type SatelliteRecord struct {
	SatelliteType string `json:"satellitetype"` // Planet/Moon

	DiscoveredBy string
	SystemZone string
	Atmosphere string // float
	Gravity    string // float
	SurfaceArea string // float
	Biosphere string

	NotableElements []string

	//SatelliteCount int
	Satellites []string
	ParentSatellite string

	TerraNullius string
	Territories int
	TerritoriesClaimed int

	UserList     []string `json:"users"`
	LastWho      time.Time `json:"lastwho"`
}

type UserRecord struct {

}

type LocationRecord struct {

}

type ElementRecord struct {

}

type ResourceRecord struct {
	ResourceType string `json:"resourcetype"` // ore / refined /

	Recipe  RecipeRecord
	Weight  string
	ResourceTier string
	FoundOn []string
}

type SkillRecord struct {

}

type RecipeRecord struct {
	RecipeName  string
	RecipeList  []RecipeItem
}

type RecipeItem struct {
	ElementType string
	Volume      string
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
func (h *InfoDBInterface) GetAllInfoRecords(c mgo.Collection) (records []InfoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	err = c.Find(bson.M{}).All(&records)
	return records, err
}

// BackerInterface function
func (h *InfoDBInterface) GetAllInfoResourceRecords(c mgo.Collection) (records []InfoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	err = c.Find(bson.M{"recordtype": "resource"}).All(&records)
	return records, err
}

// BackerInterface function
func (h *InfoDBInterface) GetAllInfoSatelliteRecords(c mgo.Collection) (records []InfoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	err = c.Find(bson.M{"recordtype": "satellite"}).All(&records)
	return records, err
}

// BackerInterface function
func (h *InfoDBInterface) GetAllMoonRecords(c mgo.Collection) (records []InfoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	err = c.Find(bson.M{"satellite.satellitetype": "moon"}).All(&records)
	return records, err
}