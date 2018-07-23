package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"math/rand"
	"time"
)

// RecruitmentHandler struct
type RecruitmentHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	recruitmentdb  *RecruitmentDB
	recruitmentChannel string
}


// Init function
func (h *RecruitmentHandler) Init() {
	h.RegisterCommands()
	h.recruitmentdb = &RecruitmentDB{db: h.db}
	h.recruitmentChannel = h.conf.Recruitment.RecruitmentChannel
}


// RegisterCommands function
func (h *RecruitmentHandler) RegisterCommands() (err error) {
	h.registry.Register("recruitment", "Create and Manage Recruitment Ads", "new|edit|delete|info|debug")
	return nil
}

// Read function
func (h *RecruitmentHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"recruitment") {
		if h.registry.CheckPermission("recruitment", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Citizen {
				h.ParseCommand(command, s, m)
			}
		}
	}
}



// ParseCommand function
func (h *RecruitmentHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command " + command + " expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "new" {
		_, commandpayload := SplitPayload(payload)
		h.NewRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "edit" {
		_, commandpayload := SplitPayload(payload)
		h.EditRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "delete" {
		_, commandpayload := SplitPayload(payload)
		h.DeleteRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "admindelete" {
		_, commandpayload := SplitPayload(payload)
		h.AdminDeleteRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "info" {
		_, commandpayload := SplitPayload(payload)
		h.RecruitmentInfo(commandpayload, s, m)
		return
	}
	if payload[0] == "viewfor" {
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "viewfor requires a user mention!")
			return
		}
		h.RecruitmentForUser(m.Mentions[0].ID, s, m)
		return
	}
	if payload[0] == "list" {
		_, commandpayload := SplitPayload(payload)
		h.ListRecruitment(commandpayload, s, m)
		return
	}
	if payload[0] == "debug" {
		_, commandpayload := SplitPayload(payload)
		h.DebugRecruitment(commandpayload, s, m)
		return
	}
}

func (h *RecruitmentHandler) RunListings(s *discordgo.Session){

	for true {
		time.Sleep(5 * time.Minute) 
		displayRecordDB, err := h.recruitmentdb.GetAllRecruitmentDisplayDB()
		if err == nil {
			if len(displayRecordDB) == 0 {
				// Yes, we are just ignoring errors here
				h.PopulateDisplayDB() // Repopulate the list
				displayRecordDB, _ = h.recruitmentdb.GetAllRecruitmentDisplayDB() // Reassign the db records
			}

			displayRecordDB = h.ShuffleRecords(displayRecordDB) // Shuffle our database

			for _, displayRecord := range displayRecordDB {
				sendingRecord, err := h.recruitmentdb.GetRecruitmentRecordFromDB(displayRecord.RecruitmentID)
				if err == nil {
					output := "**"+sendingRecord.OrgName+"**"+ "\n\n" + sendingRecord.Description
					s.ChannelMessageSend(h.conf.Recruitment.RecruitmentChannel, output)

					sendingRecord.LastRun = time.Now()
					h.recruitmentdb.UpdateRecruitmentRecord(sendingRecord)

					time.Sleep(h.conf.Recruitment.RecruitmentTimeout * time.Minute) // Only sleep if we actually found and sent valid record
				}
				h.recruitmentdb.RemoveRecruitmentDisplayRecordFromDB(displayRecord)
			}
		}
	}
}

func (h *RecruitmentHandler) PopulateDisplayDB() (err error){
	recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		return err
	}

	for _, record := range recordlist {

		uuid, err := GetUUID()
		if err != nil {
			return err
		}
		displayrecord := RecruitmentDisplayRecord{ID: uuid, RecruitmentID: record.ID}
		h.recruitmentdb.AddRecruitmentDisplayRecordToDB(displayrecord)
	}
	return nil
}


// HelpOutput function
func (h *RecruitmentHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate){
	output := "Command usage for recruitment: \n"
	output = output + "```\n"
	output = output + "new: Create a new recruitment advertisement\n"
	output = output + "edit: update an existing recruitment ad\n"
	output = output + "delete: delete a recruitment advertisement\n"
	output = output + "info: display information about a recruitment advertisement\n"
	output = output + "viewfor: view a given users recruitment ad\n"
	output = output + "admindelete: an admin command for deleting records\n"
	output = output + "list: an admin command for listing existing recruitment ads\n"
	output = output + "debug: an admin command for retrieving debug information\n"
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}


func (h *RecruitmentHandler) NewRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !h.UserIsRecruiter(m.Author.ID, s){
		s.ChannelMessageSend(m.ChannelID, "Sorry, only recruiters may use this command. Please contact an admin for the appropriate roles!")
		return
	}

	if h.RecordExistsForUser(m.Author.ID){
		s.ChannelMessageSend(m.ChannelID, "Sorry, you may only have one recruitment record active at a time!")
		return
	}

	userprivatechannel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error initializing recruitment system.")
		return
	}

	output := ":satellite_orbital: Welcome to the Dual Universe Discord Recruiter Registration System!\n\n"
	output = output + "This system will walk you through the process of creating a recruitment ad for your organization. "
	output = output + "If you encounter any issues with this process, or need assistance, please feel free to contact a Discord staff "
	output = output + "member for assistance."

	s.ChannelMessageSend(userprivatechannel.ID, output)

	time.Sleep(time.Second * 1)

	s.ChannelMessageSend(userprivatechannel.ID, "To begin the registration process, please provide your org name: ")

	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	m.ChannelID = userprivatechannel.ID
	h.callback.Watch(h.GetOrgName, uuid, "", s, m)
	return
}


func (h *RecruitmentHandler) GetOrgName(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
		return
	}

	recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieiving record list!")
		return
	}

	for _, record := range recordlist {
		if record.OrgName == m.Content {
			s.ChannelMessageSend(m.ChannelID, "Error: A record for this organization name already exists! If you believe this is an error or need assistance, please contact an admin.")
			return
		}
	}

	s.ChannelMessageSend(m.ChannelID, "You have selected: **" + m.Content + "** , is this correct?")

	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	h.callback.Watch(h.ConfirmOrgName, uuid, m.Content, s, m)
	return
}



func (h *RecruitmentHandler) ConfirmOrgName(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		uuid, err := GetUUID()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Org name confirmed. Now please provide a description for your recruitment ad: ")
		h.callback.Watch(h.GetOrgDescription, uuid, payload, s, m)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
	return
}


func (h *RecruitmentHandler) GetOrgDescription(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
		return
	}

	output := "The following description was provided (please confirm with Y/N): \n" + m.Content
	s.ChannelMessageSend(m.ChannelID, output)

	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}

	callbackPayload := payload + "|||||" + m.Content
	h.callback.Watch(h.ConfirmOrgDescription, uuid, callbackPayload, s, m)
	return
}


func (h *RecruitmentHandler) ConfirmOrgDescription(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		uuid, err := GetUUID()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
			return
		}

		splitPayload := strings.Split(payload, "|||||")
		output := "**"+splitPayload[0]+"**" + "\n\n"
		output = output + splitPayload[1]

		s.ChannelMessageSend(m.ChannelID, "Your recruitment ad will be displayed as follows (please confirm with Y/N)\n ---- ")
		time.Sleep(time.Second*1)
		s.ChannelMessageSend(m.ChannelID, output)

		h.callback.Watch(h.ConfirmRecruitmentAd, uuid, payload, s, m)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
	return
}

func (h *RecruitmentHandler) ConfirmRecruitmentAd(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {

		uuid, err := GetUUID()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
			return
		}
		splitPayload := strings.Split(payload, "|||||")
		output := "**"+splitPayload[0]+"**" + "\n"
		output = output + splitPayload[1]

		recruitmentAd := RecruitmentRecord{ID: uuid, OwnerID: m.Author.ID, OrgName: splitPayload[0], Description: splitPayload[1]}

		err = h.recruitmentdb.AddRecruitmentRecordToDB(recruitmentAd)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "There was an error processing your request, please contact an admin!")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Your advertisement has been saved successfully!")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
	return
}


func (h *RecruitmentHandler) UserIsRecruiter(userID string, s *discordgo.Session) (bool) {

	member, err := s.GuildMember(h.conf.DiscordConfig.GuildID, userID)
	if err  != nil {
		return false
	}
	for _, roleID := range member.Roles {

		rolename, err := getRoleNameByID(roleID, h.conf.DiscordConfig.GuildID, s)
		if err != nil {
			return false
		}

		if rolename == "Recruiter" {
			return true
		}
	}
	return false
}

func (h *RecruitmentHandler) RecordExistsForUser(userID string) (bool) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", userID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + userID)
		return false
	}

	recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		return false
	}
	for _, record := range recordlist {

		if record.OwnerID == userID {
			return true
		}
	}
	return false
}

func (h *RecruitmentHandler) EditRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	s.ChannelMessageSend(m.ChannelID, "This command is under construction, please delete your ad and re-create it until this has been implemented!")
	return
}

func (h *RecruitmentHandler) DeleteRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !h.RecordExistsForUser(m.Author.ID){
		s.ChannelMessageSend(m.ChannelID, "Sorry, no recruitment record was found for your user ID. If you believe this is an error, please contact an admin!")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Are you sure you would like to delete your recruitment ad? (Y/N)")

	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	h.callback.Watch(h.ConfirmDeleteRecruitmentAd, uuid, "", s, m)
	return
}

func (h *RecruitmentHandler) ConfirmDeleteRecruitmentAd(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {
		recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieiving record list!")
			return
		}
		for _, record := range recordlist {
			if record.OwnerID == m.Author.ID {
				h.recruitmentdb.RemoveRecruitmentRecordFromDB(record)
				s.ChannelMessageSend(m.ChannelID, "Your recruitment advertisement has been removed successfully.")
				return
			}
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Recruitment ad deletion cancelled.")
	return
}


func (h *RecruitmentHandler) AdminDeleteRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Admin {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Command 'admindelete' expects an argument (the ID of the record to be removed)")
		return
	}
	err = h.recruitmentdb.RemoveRecruitmentRecordFromDBByID(payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error removing record: " + err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Record removed successfully!")
	return

}

func (h *RecruitmentHandler) RecruitmentInfo(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Admin {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "info requires an argument!")
		return
	}

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
		return
	}

	output := "Record info:\n```\n"
	output = output + "ID: " + record.ID + "\n"

	userrecord, err := s.User(record.OwnerID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving user: " + err.Error())
		return
	}
	output = output + "Owner: " + userrecord.Username + "\n"
	output = output + "Owner ID: " + record.OwnerID + "\n"
	output = output + "Last Run: " + record.LastRun.Format("2006-01-02 15:04:05") + "\n"
	output = output + "Org Name: " + record.OrgName + "\n"
	output = output + "Description: " + record.Description + "\n"
	output = output + "\n```\n"
	s.ChannelMessageSend(m.ChannelID, output)
	return
}

func (h *RecruitmentHandler) RecruitmentForUser(userID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Admin {
		return
	}

	records, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
		return
	}

	for _, record := range records {
		if record.OwnerID == userID {
			output := "Record info:\n```\n"
			output = output + "ID: " + record.ID + "\n"

			userrecord, err := s.User(record.OwnerID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving discord user: " + err.Error())
				return
			}
			output = output + "Owner: " + userrecord.Username + "\n"
			output = output + "Owner ID: " + record.OwnerID + "\n"
			output = output + "Last Run: " + record.LastRun.Format("2006-01-02 15:04:05") + "\n"
			output = output + "Org Name: " + record.OrgName + "\n"
			output = output + "Description: " + record.Description + "\n"
			output = output + "\n```\n"
			s.ChannelMessageSend(m.ChannelID, output)
			return
		}
	}

	s.ChannelMessageSend(m.ChannelID, "Error: No recruitment record found for user!")
	return
}

func (h *RecruitmentHandler) ListRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Admin {
		return
	}

	recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieiving record list!")
		return
	}
	output := ":satellite: Recruitment records: \n```\n"

	for _, record := range recordlist {
		userrecord, err := s.User(record.OwnerID)
		if err == nil {
			output = output + record.OrgName + " - " + userrecord.Username + " - "+ record.ID + "\n"
		}
	}
	output = output + "\n```\n"
	s.ChannelMessageSend(m.ChannelID, output)
	return
}


func (h *RecruitmentHandler) DebugRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Admin {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Debug requires an argument!")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Command under construction.")
	return
}

// ShuffleRecords function
// We use this for shuffling our record list every iteration so we don't lose records on a bot restart
func (h *RecruitmentHandler) ShuffleRecords(DisplayRecords []RecruitmentDisplayRecord) (ShuffledRecords []RecruitmentDisplayRecord){

	for i := len(DisplayRecords) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		DisplayRecords[i], DisplayRecords[j] = DisplayRecords[j], DisplayRecords[i]
	}

	return DisplayRecords
}