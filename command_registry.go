package main


/*
The command registry adds the ability for commands to be "protected" by permissions.

 Now I use quotes because there are no guarantees that a command may inadvertently kick off a process if permissions aren't checked properly.

 */

import (
	"errors"
)

type CommandRegistry struct {

	db *DBHandler

}


type CommandRecord struct {

	ID string `storm:"id"`
	Command string `storm:"index"`
	Groups []string `storm:"index"`
	Channels []string `storm:"index"`
	Users []string `storm:"index"`

}

func (h *CommandRegistry) CreateCommand(command string) (err error) {

	db := h.db.rawdb.From("Commands")

	record := CommandRecord{}
	err = db.Find("Command", command, &record)
	if err == nil {
		return errors.New("Command already exists!")
	}

	record.Command = command

	err = h.SaveCommand(record)
	if err != nil {
		return err
	}

	return nil
}

func (h *CommandRegistry) SaveCommand(record CommandRecord) (err error) {
	db := h.db.rawdb.From("Commands")
	err = db.Save(&record)
	return err
}

func (h *CommandRegistry) RemoveCommand(record CommandRecord) (err error) {

	db := h.db.rawdb.From("Commands")
	err = db.Remove(&record)
	return err
}

func (h *CommandRegistry) GetCommand(command string) (record CommandRecord, err error){

	db := h.db.rawdb.From("Commands")

	err = db.Find("Command", command, &record)
	if err != nil {
		return record, err
	}

	return record, nil
}

func (h *CommandRegistry) AddGroup(command string, group string) (err error){

	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	for _, v := range record.Groups {
		if v == group {
			return errors.New("Command already belongs to group " + group)
		}
	}

	record.Groups = append(record.Groups, group)
	return nil

}

func (h *CommandRegistry) RemoveGroup(command string, group string) (err error){

	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	for _, v := range record.Groups {
		if v == group {
			record.Groups = RemoveStringFromSlice(record.Groups, group)
			h.SaveCommand(record)
			return nil
		}
	}

	return errors.New("Command does not belong to group " + group)
}

func (h *CommandRegistry) GetGroups(command string) (groups []string, err error){

	record, err := h.GetCommand(command)
	if err != nil {
		return groups, err
	}

	return record.Groups, nil
}


func (h *CommandRegistry) AddChannel(command string, channel string) (err error){

	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	for _, v := range record.Channels {
		if v == channel {
			return errors.New("Command already belongs to channel " + channel)
		}
	}

	record.Channels = append(record.Channels, channel)
	return nil
}

func (h *CommandRegistry) RemoveChannel(command string, channel string) (err error){
	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	for _, v := range record.Channels {
		if v == channel {
			record.Channels = RemoveStringFromSlice(record.Channels, channel)
			h.SaveCommand(record)
			return nil
		}
	}

	return errors.New("Command does not belong to channel " + channel)
}

func (h *CommandRegistry) GetChannels(command string) (channels []string, err error) {

	record, err := h.GetCommand(command)
	if err != nil {
		return channels, err
	}

	return record.Channels, nil
}


func (h *CommandRegistry) CheckChannel(command string, channel string) (bool) {

	channels, err := h.GetChannels(command)
	if err != nil {
		return false
	}

	for _, c := range channels {
		if c == channel {
			return true
		}
	}

	return false
}


func (h *CommandRegistry) CheckGroup(command string, group string) (bool) {

	groups, err := h.GetGroups(command)
	if err != nil {
		return false
	}

	for _, c := range groups {
		if c == group {
			return true
		}
	}

	return false
}


func (h *CommandRegistry) AddUser(command string, user string) (err error){

	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	for _, v := range record.Users {
		if v == user {
			return errors.New(user + " already has permission to use " + command)
		}
	}

	record.Users = append(record.Users, user)
	return nil
}

func (h *CommandRegistry) RemoveUser(command string, user string) (err error){
	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	for _, v := range record.Users {
		if v == user {
			record.Users = RemoveStringFromSlice(record.Users, user)
			h.SaveCommand(record)
			return nil
		}
	}

	return errors.New(user + " does not have permission to use " + command)
}


func (h *CommandRegistry) GetUsers(command string) (users []string, err error) {

	record, err := h.GetCommand(command)
	if err != nil {
		return users, err
	}

	return record.Users, nil
}


func (h *CommandRegistry) CheckUser(command string, user string) (bool) {

	users, err := h.GetUsers(command)
	if err != nil {
		return false
	}

	for _, c := range users {
		if c == user {
			return true
		}
	}

	return false
}

func (h *CommandRegistry) CheckPermission(command string, channel string, user User) (bool) {

	if h.CheckUser(command, user.ID){
		return true
	}
	if h.CheckChannel(command, channel){
		return true
	}

	// Check Groups
	cmd, err := h.GetCommand(command)
	if err != nil{
		return false
	}

	for _, group := range cmd.Groups  {
		if user.CheckRole(group) {
			return true
		}
	}

	return false

}