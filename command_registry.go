package main


/*
The command registry adds the ability for commands to be "protected" by permissions.

 Now I use quotes because there are no guarantees that a command may inadvertently kick off a process if permissions aren't checked properly.

 */

import (
	"errors"
	"fmt"
)

type CommandRegistry struct {

	db *DBHandler
	conf *Config

}


type CommandRecord struct {

	Command string `storm:"id"`
	Groups []string `storm:"index"`
	Channels []string `storm:"index"`
	Users []string `storm:"index"`
	Description string
	Usage string

}

// Does the same thing as Create Command without a return value
func (h *CommandRegistry) Register(command string, description string, usage string) {

	db := h.db.rawdb.From("Commands")

	record := CommandRecord{}
	err := db.One("Command", command, &record)
	if err == nil {
		return // command already exists
	}

	record.Command = command
	record.Description = description
	record.Usage = usage

	err = h.SaveCommand(record)
	if err != nil {
		fmt.Println("Could not save command to registry" + err.Error())
		return
	}

	return
}

func (h *CommandRegistry) CreateCommand(command string) (err error) {

	db := h.db.rawdb.From("Commands")

	record := CommandRecord{}
	err = db.One("Command", command, &record)
	if err == nil {
		return errors.New("Command already exists")
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
	err = db.DeleteStruct(&record)
	err = db.Save(&record)
	return err
}

func (h *CommandRegistry) RemoveCommand(record CommandRecord) (err error) {

	db := h.db.rawdb.From("Commands")
	err = db.DeleteStruct(&record)
	return err
}

func (h *CommandRegistry) GetCommand(command string) (record CommandRecord, err error){

	db := h.db.rawdb.From("Commands")
	record = CommandRecord{}

	err = db.One("Command", command, &record)
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
	h.SaveCommand(record)
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
	h.SaveCommand(record)
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
	h.SaveCommand(record)
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

func (h *CommandRegistry) ChannelList(command string) (channels []string, err error){
	// Check Channels
	cmd, err := h.GetCommand(command)
	if err != nil{
		fmt.Println("Could not get command for channel list")
		return channels, err
	}

	for _, channel := range cmd.Channels  {
		channels = append(channels, channel)
	}
	return channels, nil
}

func (h *CommandRegistry) UserList(command string) (users []string, err error){
	// Check Users
	cmd, err := h.GetCommand(command)
	if err != nil{
		return users, err
	}

	for _, user := range cmd.Users  {
		users = append(users, user)
	}
	return users, nil
}

func (h *CommandRegistry) GroupList(command string) (groups []string, err error){

	// Check Groups
	cmd, err := h.GetCommand(command)
	if err != nil{
		return groups, err
	}

	for _, group := range cmd.Groups  {
		groups = append(groups, group)
	}
	return groups, nil
}

func (h *CommandRegistry) CommandsForChannel(page int, channel string) (records []CommandRecord, err error){

	db := h.db.rawdb.From("Commands")

	// Sanitize our page count
	pagecount, err := h.CommandsForChannelPageCount(channel)
	if err != nil{
		return records, err
	}
	if page > pagecount {
		page = pagecount
	}
	if page < 0 {
		return records, errors.New("Invalid page")
	}

	commandrecords := []CommandRecord{}
	err = db.All(&commandrecords)
	if err != nil{
		return records, err
	}

	recordcount := 0
	currentrecordcount := 0
	for _, record := range commandrecords  {
			for _, channelID := range record.Channels {
				if channelID == channel {
					currentrecordcount = currentrecordcount + 1
					if currentrecordcount > page * h.conf.DUBotConfig.PerPageCount {
						if recordcount < h.conf.DUBotConfig.PerPageCount {
							records = append(records, record)
							recordcount = recordcount + 1
						}
					}
				}
			}
		}


	if len(records) < 1{
		return records, errors.New("not found")
	}

	return records, nil
}

func (h *CommandRegistry) CommandsForChannelCount(channel string) (count int, err error) {
	db := h.db.rawdb.From("Commands")

	commandrecords := []CommandRecord{}
	err = db.All(&commandrecords)
	if err != nil{
		return 0, err
	}
	for _, record := range commandrecords  {
		for _, channelID:= range record.Channels {
			if channelID == channel {
				count = count + 1
			}
		}
	}
	return count, nil
}



func (h *CommandRegistry) CommandsForChannelPageCount(channel string) (pages int, err error){
	// Check Groups
	db := h.db.rawdb.From("Commands")

	pages = 0

	commandrecords := []CommandRecord{}
	err = db.All(&commandrecords)
	if err != nil{
		return pages, err
	}

	commandCount, err := h.CommandsForChannelCount(channel)
	if err != nil {
		return pages, err
	}

	if commandCount < 1{
		return pages, errors.New("not found")
	}

	pages = (commandCount / h.conf.DUBotConfig.PerPageCount) + 1

	return pages, nil
}



func (h *CommandRegistry) SetDescription(command string, description string) (err error) {

	record, err := h.GetCommand(command)
	if err != nil{
		return err
	}

	record.Description = description

	err = h.SaveCommand(record)
	if err != nil{
		return err
	}

	return nil
}

func (h *CommandRegistry) GetDescription(command string) (description string, err error) {

	record, err := h.GetCommand(command)
	if err != nil {
		return description, err
	}

	description = record.Description
	return description, nil
}

func (h *CommandRegistry) SetUsage(command string, usage string) (err error) {

	record, err := h.GetCommand(command)
	if err != nil {
		return err
	}

	record.Usage = usage

	err = h.SaveCommand(record)
	if err != nil{
		return err
	}

	return nil
}

func (h *CommandRegistry) GetUsage(command string) (usage string, err error) {

	record, err := h.GetCommand(command)
	if err != nil {
		return usage, err
	}

	usage = record.Usage
	return usage, nil
}