package main

import (
	"errors"
)

/*
Channels as tracked by the permissions system

Permissions
Citizens - Default true (be aware of this)
Agora - Agora Role
Newsroom - Editors role
Recruitment - Recruiters role
Streamers - Streamers role

Admin - Admins only channel
Staff - Staff only channel
Moderators - Moderators and Senior Moderators
Senior Mods - Admins and Senior Mods

Bot Log - Logger callbacks go here
Promotion Log - Promotion and Demotion logs
Bank Log - Banking Logs

HQ - Accept all commands in this room
*/

// ChannelDB struct
type ChannelDB struct {
	db *DBHandler
}

// ChannelRecord struct
type ChannelRecord struct {
	ID     string `storm:"id"`
	Groups []string

	IsBotLog        bool
	IsPermissionLog bool
	IsBankLog       bool
	IsMusicRoom     bool
	IsMusicAudio    bool

	HQ bool
}

// CreateChannel function
func (h *ChannelDB) CreateChannel(channelid string) (err error) {

	db := h.db.rawdb.From("Channels")

	record := ChannelRecord{}
	err = db.One("ID", channelid, &record)
	if err == nil {
		return errors.New("Channel Record already exists")
	}

	record.ID = channelid

	err = h.SaveChannel(record)
	if err != nil {
		return err
	}

	return nil
}

// GetChannel function
func (h *ChannelDB) GetChannel(channelid string) (record ChannelRecord, err error) {
	db := h.db.rawdb.From("Channels")
	record = ChannelRecord{}

	err = db.One("ID", channelid, &record)
	if err != nil {
		return record, err
	}
	return record, nil
}

// SaveChannel function
func (h *ChannelDB) SaveChannel(record ChannelRecord) (err error) {
	db := h.db.rawdb.From("Channels")
	err = db.DeleteStruct(&record)
	err = db.Save(&record)
	return err
}

// RemoveChannel function
func (h *ChannelDB) RemoveChannel(record ChannelRecord) (err error) {
	db := h.db.rawdb.From("Channels")
	err = db.DeleteStruct(&record)
	return err
}

// GetDB function
func (h *ChannelDB) GetDB() (records []ChannelRecord, err error) {
	commandrecords := []ChannelRecord{}
	db := h.db.rawdb.From("Channels")

	err = db.All(&commandrecords)
	if err != nil {
		return records, err
	}
	return commandrecords, nil
}

// CreateIfNotExists function
func (h *ChannelDB) CreateIfNotExists(channelid string) (err error) {

	err = h.CreateChannel(channelid)
	if err != nil {
		if err.Error() == "Channel Record already exists" {
			return nil
		}
		return err
	}
	return nil
}

// AddGroup function
func (h *ChannelDB) AddGroup(channelid string, group string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}
	record, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	for _, v := range record.Groups {
		if v == group {
			return errors.New("Channel already belongs to group " + group)
		}
	}

	record.Groups = append(record.Groups, group)
	err = h.SaveChannel(record)
	if err != nil {
		return err
	}

	return nil
}

// RemoveGroup function
func (h *ChannelDB) RemoveGroup(channelid string, group string) (err error) {

	record, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	for _, v := range record.Groups {
		if v == group {
			record.Groups = RemoveStringFromSlice(record.Groups, group)
			h.SaveChannel(record)
			return nil
		}
	}

	return errors.New("Channel does not belong to group " + group)
}

// GetGroups function
func (h *ChannelDB) GetGroups(channelid string) (groups []string, err error) {

	record, err := h.GetChannel(channelid)
	if err != nil {
		return groups, err
	}

	return record.Groups, nil
}

// SetBotLog function
func (h *ChannelDB) SetBotLog(channelid string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsBotLog {
			return errors.New("Botlog already assigned")
		}
	}

	channelrecord, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	channelrecord.IsBotLog = true
	err = h.SaveChannel(channelrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetBotLog function
func (h *ChannelDB) GetBotLog() (channelid string, err error) {
	channelrecords, err := h.GetDB()
	if err != nil {
		return "", err
	}

	for _, record := range channelrecords {
		if record.IsBotLog {
			return record.ID, nil
		}
	}
	return "", errors.New("Bot Log Channel Not Found")
}

// RemoveBotLog function
func (h *ChannelDB) RemoveBotLog() (err error) {

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsBotLog {
			record.IsBotLog = false
			err := h.SaveChannel(record)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("Bot Log Channel Not Found")
}

// SetPermissionLog function
func (h *ChannelDB) SetPermissionLog(channelid string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsPermissionLog {
			return errors.New("Permission Log already assigned")
		}
	}

	channelrecord, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	channelrecord.IsPermissionLog = true
	err = h.SaveChannel(channelrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetPermissionLog function
func (h *ChannelDB) GetPermissionLog() (channelid string, err error) {
	channelrecords, err := h.GetDB()
	if err != nil {
		return "", err
	}

	for _, record := range channelrecords {
		if record.IsPermissionLog {
			return record.ID, nil
		}
	}
	return "", errors.New("Permission Log Channel Not Found")
}

// RemovePermissionLog function
func (h *ChannelDB) RemovePermissionLog() (err error) {

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsPermissionLog {
			record.IsPermissionLog = false
			err := h.SaveChannel(record)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("Permission Log Channel Not Found")
}

// SetBankLog function
func (h *ChannelDB) SetBankLog(channelid string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsBankLog {
			return errors.New("Bank Log already assigned")
		}
	}

	channelrecord, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	channelrecord.IsBankLog = true
	err = h.SaveChannel(channelrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetBankLog function
func (h *ChannelDB) GetBankLog() (channelid string, err error) {
	channelrecords, err := h.GetDB()
	if err != nil {
		return "", err
	}

	for _, record := range channelrecords {
		if record.IsBankLog {
			return record.ID, nil
		}
	}
	return "", errors.New("Bank Log Channel Not Found")
}

// RemoveBankLog function
func (h *ChannelDB) RemoveBankLog() (err error) {

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsBankLog {
			record.IsBankLog = false
			err := h.SaveChannel(record)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("Bank Log Channel Not Found")
}

// SetHQ function
func (h *ChannelDB) SetHQ(channelid string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.HQ {
			return errors.New("HQ already assigned")
		}
	}

	channelrecord, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	channelrecord.HQ = true
	err = h.SaveChannel(channelrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetHQ function
func (h *ChannelDB) GetHQ() (channelid string, err error) {
	channelrecords, err := h.GetDB()
	if err != nil {
		return "", err
	}

	for _, record := range channelrecords {
		if record.HQ {
			return record.ID, nil
		}
	}
	return "", errors.New("HQ Channel Not Found")
}

// RemoveHQ function
func (h *ChannelDB) RemoveHQ() (err error) {

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.HQ {
			record.HQ = false
			err := h.SaveChannel(record)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("HQ Channel Not Found")
}

// SetMusicRoom function
func (h *ChannelDB) SetMusicRoom(channelid string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsMusicRoom {
			return errors.New("MusicRoom already assigned")
		}
	}

	channelrecord, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	channelrecord.IsMusicRoom = true
	err = h.SaveChannel(channelrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetMusicRoom function
func (h *ChannelDB) GetMusicRoom() (channelid string, err error) {
	channelrecords, err := h.GetDB()
	if err != nil {
		return "", err
	}

	for _, record := range channelrecords {
		if record.IsMusicRoom {
			return record.ID, nil
		}
	}
	return "", errors.New("Music Room Not Found")
}

// RemoveMusicRoom function
func (h *ChannelDB) RemoveMusicRoom() (err error) {

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsMusicRoom {
			record.IsMusicRoom = false
			err := h.SaveChannel(record)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("Music Room Not Found")
}

// SetMusicAudio function
func (h *ChannelDB) SetMusicAudio(channelid string) (err error) {

	err = h.CreateIfNotExists(channelid)
	if err != nil {
		return err
	}

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsMusicAudio {
			return errors.New("MusicAudio already assigned")
		}
	}

	channelrecord, err := h.GetChannel(channelid)
	if err != nil {
		return err
	}

	channelrecord.IsMusicAudio = true
	err = h.SaveChannel(channelrecord)
	if err != nil {
		return err
	}

	return nil
}

// GetMusicAudio function
func (h *ChannelDB) GetMusicAudio() (channelid string, err error) {
	channelrecords, err := h.GetDB()
	if err != nil {
		return "", err
	}

	for _, record := range channelrecords {
		if record.IsMusicAudio {
			return record.ID, nil
		}
	}
	return "", errors.New("Music Audio Channel Not Found")
}

// RemoveMusicAudio function
func (h *ChannelDB) RemoveMusicAudio() (err error) {

	channelrecords, err := h.GetDB()
	if err != nil {
		return err
	}

	for _, record := range channelrecords {
		if record.IsMusicAudio {
			record.IsMusicAudio = false
			err := h.SaveChannel(record)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("Music Audio Channel Not Found")
}
