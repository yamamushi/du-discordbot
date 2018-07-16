package main

import (
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
	ID      string `storm:"id"`
	OwnerID  string
	OrgName  string
	Description string
	LastRun  time.Time
}

// We store our records in the list so that bot reboots don't break displaying them
type RecruitmentDisplayRecord struct {

	ID      string `storm:"id"`
	RecruitmentID string

}

