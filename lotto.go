package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

// LottoDB struct
type LottoDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// LottoRecord struct
type LottoRecord struct {
	ID      string `storm:"id"`
	OwnerID  string
	ShortName string
	Description string
	CreatedDate    time.Time
	Duration time.Duration
	Completed bool
}

// LottoEntry struct
type LottoEntry struct {
	ID      string `storm:"id"`
	LottoID string `storm:"id"`
	UserID  string
	Date    time.Time
}


// AddLottoToDB function
func (h *LottoDB) AddLottoRecordToDB(record LottoRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("LottoDB")
	err = db.Save(&record)
	return err
}

// RemoveLottoFromDB function
func (h *LottoDB) RemoveLottoRecordFromDB(record LottoRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("LottoDB")
	err = db.DeleteStruct(&record)
	return err
}

// RemoveLottoFromDBByID function
func (h *LottoDB) RemoveLottoRecordFromDBByID(lottoID string, s *discordgo.Session) (err error) {

	Lotto, err := h.GetLottoFromDB(lottoID)
	if err != nil {
		return err
	}

	err = h.RemoveLottoRecordFromDB(Lotto)
	if err != nil {
		return err
	}

	return nil
}

// GetLottoFromDB function
func (h *LottoDB) GetLottoFromDB(lottoID string) (Record LottoRecord, err error) {

	LottoDB, err := h.GetAllLottoDB()
	if err != nil {
		return Record, err
	}

	for _, i := range LottoDB {
		if i.ID == lottoID {
			return i, nil
		}
	}

	return Record, errors.New("No record found")
}

// GetAllLottoDB function
func (h *LottoDB) GetAllLottoDB() (RecordList []LottoRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("LottoDB")
	err = db.All(&RecordList)
	if err != nil {
		return RecordList, err
	}

	return RecordList, nil
}
