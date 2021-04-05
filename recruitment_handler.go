package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RecruitmentHandler struct
type RecruitmentHandler struct {
	conf               *Config
	registry           *CommandRegistry
	callback           *CallbackHandler
	db                 *DBHandler
	reactions          *ReactionsHandler
	recruitmentdb      *RecruitmentDB
	recruitmentChannel string
	userdb             *UserHandler
	globalstate        *StateDB
	configdb           *ConfigDB
	timeoutchan        chan bool
	querylocker        sync.RWMutex
	lastpost           time.Time
}

// Init function
func (h *RecruitmentHandler) Init() {
	h.RegisterCommands()
	h.recruitmentdb = &RecruitmentDB{db: h.db}
	h.recruitmentChannel = h.conf.Recruitment.RecruitmentChannel
	h.timeoutchan = make(chan bool)
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
		s.ChannelMessageSend(m.ChannelID, "Command "+command+" expects an argument, see help for usage.")
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
	if payload[0] == "setlimit" {
		_, commandpayload := SplitPayload(payload)
		h.SetUserRecordLimit(commandpayload, s, m)
		return
	}
	if payload[0] == "forcepost" {
		_, commandpayload := SplitPayload(payload)
		h.ForcePost(commandpayload, s, m)
		return
	}
	if payload[0] == "info" {
		_, commandpayload := SplitPayload(payload)
		h.RecruitmentInfo(commandpayload, s, m)
		return
	}
	if payload[0] == "validate" {
		_, commandpayload := SplitPayload(payload)
		h.Validate(commandpayload, s, m)
		return
	}
	if payload[0] == "queue" {
		_, commandpayload := SplitPayload(payload)
		h.QueueInfo(commandpayload, s, m)
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
	if payload[0] == "approve" || payload[0] == "deny" || payload[0] == "reject" {
		command, commandpayload := SplitPayload(payload)
		h.ModifyApproval(command, commandpayload, s, m)
		return
	}
	if payload[0] == "fixusers" {
		//_, commandpayload := SplitPayload(payload)
		//err := h.FixUsers()
		//if err != nil {
		//s.ChannelMessageSend(m.ChannelID, "Error: This command is disabled." )
		return
		//}
		//s.ChannelMessageSend(m.ChannelID, "User recruitment records updated")
		//return
	}
	if payload[0] == "fixads" {
		/*
			err := h.FixAds()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error() )
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Recruitment records repaired")
		*/
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Unrecognized option: "+payload[0])
	return
}

func (h *RecruitmentHandler) RunListings(s *discordgo.Session) {

	if h.conf.Recruitment.RecruitmentWaitOnStartup {

		select {
		case <-h.timeoutchan:
			break
		case <-time.After(5 * time.Minute):
			break
		}
	}

	for true {
		displayRecordDB, err := h.recruitmentdb.GetAllRecruitmentDisplayDB()
		if err == nil {
			if len(displayRecordDB) == 0 {
				// Yes, we are just ignoring errors here
				h.PopulateDisplayDB()                                             // Repopulate the list
				displayRecordDB, _ = h.recruitmentdb.GetAllRecruitmentDisplayDB() // Reassign the db records
			}

			displayRecordDB = h.ShuffleRecords(displayRecordDB) // Shuffle our database

			for _, displayRecord := range displayRecordDB {
				h.recruitmentdb.RemoveRecruitmentDisplayRecordFromDB(displayRecord) // We remove the record here to avoid conflicts on bot restarts.

				sendingRecord, err := h.recruitmentdb.GetRecruitmentRecordFromDB(displayRecord.RecruitmentID)
				if err == nil {
					globalstate, err := h.globalstate.GetState()
					if err == nil {
						if sendingRecord.ID != globalstate.LastRecruitmentIDPosted {
							output := "**" + sendingRecord.OrgName + "**" + "\n\n" + sendingRecord.Description

							if sendingRecord.Approved {
								s.ChannelMessageSend(h.conf.Recruitment.RecruitmentChannel, output)
								sendingRecord.LastRun = time.Now()
								h.lastpost = time.Now()
								h.recruitmentdb.UpdateRecruitmentRecord(sendingRecord)

								globalstate.LastRecruitmentIDPosted = sendingRecord.ID
								h.globalstate.SetState(globalstate)

								timercount, err := h.configdb.GetValue("recruitment-timer")
								if err != nil {
									timercount = int(h.conf.Recruitment.RecruitmentTimeout)
								}
								if timercount == 0 {
									timercount = 1
								}

								// Sleep or wait
								select {
								case <-h.timeoutchan:
									continue
								case <-time.After(time.Duration(timercount) * time.Minute):
									break
								}
							}
						}
					}
				}
			}
		}
	}
}

func (h *RecruitmentHandler) CheckValidity(s *discordgo.Session) {
	/*
		if h.conf.Recruitment.RecruitmentWaitOnStartup {

			select {
			case <-h.timeoutchan:
				break
			case <-time.After(5 * time.Minute):
				break
			}
		}
	*/

	for true {
		time.Sleep(5 * time.Minute)

		recruitmentRecords, err := h.recruitmentdb.GetAllRecruitmentDB()
		if err == nil {
			for _, record := range recruitmentRecords {

				expirationTime, err := h.configdb.GetValue("recruitment-expiration")
				if err != nil {
					expirationTime = int(h.conf.Recruitment.RecruitmentExpiration)
				}

				expiryMinutes := time.Duration(time.Duration(expirationTime) * time.Minute)
				expiryDate := record.LastValidated.Add(expiryMinutes)

				if time.Now().After(expiryDate) && record.ValidationReminderSent {
					h.NotifyOwnerExpiredRecord(record.ID, s)
					h.recruitmentdb.RemoveRecruitmentRecordFromDBByID(record.ID)
				}

				reminderTime, err := h.configdb.GetValue("recruitment-reminder")
				if err != nil {
					expirationTime = int(h.conf.Recruitment.RecruitmentExpiration)
				}
				reminderMinutes := time.Duration(time.Duration(expirationTime)*time.Minute) - time.Duration(time.Duration(reminderTime)*time.Minute)
				reminderDate := record.LastValidated.Add(reminderMinutes)

				if time.Now().After(reminderDate) && !record.ValidationReminderSent {
					record.ValidationReminderSent = true
					err = h.recruitmentdb.UpdateRecruitmentRecord(record)
					if err == nil {
						h.VerifyValidityWithOwner(record.ID, s)
					}
				}
			}
		}
	}
}

func (h *RecruitmentHandler) NotifyOwnerExpiredRecord(recordID string, s *discordgo.Session) {

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(recordID)
	if err != nil {
		return
	}

	userprivatechannel, err := s.UserChannelCreate(record.OwnerID)
	if err != nil {
		return
	}

	output := ":rotating_light: This is a notification that your recruitment advertisement for \"**" + record.OrgName + "**\" has expired. \n\n" +
		"This is done on an automated basis to keep the recruitment system from being filled with abandoned advertisements. " +
		"Feel free to create a new advertisement at any time."

	s.ChannelMessageSend(userprivatechannel.ID, output)

	unformattedAdminChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
	if err != nil {
		return
	}

	formattedAdminChannel := CleanChannel(unformattedAdminChannel)
	s.ChannelMessageSend(formattedAdminChannel, "Recruitment ad for \"**"+record.OrgName+"**\" has expired and has been removed automatically.")
	return
}

func (h *RecruitmentHandler) VerifyValidityWithOwner(recordID string, s *discordgo.Session) {

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(recordID)
	if err != nil {
		return
	}

	userprivatechannel, err := s.UserChannelCreate(record.OwnerID)
	if err != nil {
		return
	}

	output := ":rotating_light: This is a reminder that your recruitment advertisement for \"**" + record.OrgName + "**\" is about to expire in **72 hours**. \n\n" +
		"This is done on an automated basis to keep the recruitment system from being filled with abandoned advertisements.\n\n" +
		"If you would like to prevent your advertisement from being removed automatically, please use the following command in the Dual Universe discord:" +
		"\n```\n~recruitment validate " + record.ID + "\n```\n"

	s.ChannelMessageSend(userprivatechannel.ID, output)
}

func (h *RecruitmentHandler) Validate(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "validate requires an argument.")
		return
	}

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	user, err := h.userdb.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if m.Author.ID != record.OwnerID && !user.Moderator {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to validate this advertisement. If you believe this is an error, pelase contact a moderator.")
		return
	}

	expirationTime, err := h.configdb.GetValue("recruitment-expiration")
	if err != nil {
		expirationTime = int(h.conf.Recruitment.RecruitmentExpiration)
	}

	expiryMinutes := time.Duration(time.Duration(expirationTime) * time.Minute)
	expiryDate := record.LastValidated.Add(expiryMinutes)

	reminderTime, err := h.configdb.GetValue("recruitment-reminder")
	if err != nil {
		expirationTime = int(h.conf.Recruitment.RecruitmentExpiration)
	}
	reminderMinutes := time.Duration(time.Duration(expirationTime)*time.Minute) - time.Duration(time.Duration(reminderTime)*time.Minute)
	reminderDate := record.LastValidated.Add(reminderMinutes)

	if !time.Now().After(reminderDate) {
		s.ChannelMessageSend(m.ChannelID, "This command is only valid for an expiring advertisement.")
		return
	}

	record.LastValidated = time.Now()
	record.ValidationReminderSent = false
	h.recruitmentdb.UpdateRecruitmentRecord(record)

	expiryDate = record.LastValidated.Add(expiryMinutes)
	loc, _ := time.LoadLocation("America/Chicago")
	s.ChannelMessageSend(m.ChannelID, "Advertisement has been successfully validated until "+expiryDate.In(loc).Format("Mon Jan _2 03:04 MST 2006"))

	unformattedAdminChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
	if err != nil {
		return
	}

	formattedAdminChannel := CleanChannel(unformattedAdminChannel)

	s.ChannelMessageSend(formattedAdminChannel, "Recruitment ad for \"**"+record.OrgName+"**\" has been validated by **"+m.Author.Username+"**("+m.Author.ID+") and will not be removed automatically until the next expiration period.")

}

func (h *RecruitmentHandler) PopulateDisplayDB() (err error) {
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
func (h *RecruitmentHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate) {
	output := "Command usage for recruitment: \n"
	output = output + "```\n"
	output = output + "new: Create a new recruitment advertisement\n"
	output = output + "edit: update an existing recruitment ad\n"
	output = output + "delete: delete a recruitment advertisement\n"
	output = output + "info: display information about a recruitment advertisement\n"
	output = output + "viewfor: view a given users recruitment ad\n"
	output = output + "admindelete: an admin command for deleting records\n"
	output = output + "setlimit: sets a users record limit\n"
	output = output + "list: an admin command for listing existing recruitment ads\n"
	output = output + "debug: an admin command for retrieving debug information\n"
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}

func (h *RecruitmentHandler) NewRecruitment(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !h.UserIsRecruiter(m.Author.ID, s) {
		s.ChannelMessageSend(m.ChannelID, "Sorry, only recruiters may use this command. Please contact an admin for the appropriate roles!")
		return
	}

	recordlimit, err := h.GetUserRecordLimit(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	recordcount, err := h.GetUserRecordCount(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if recordlimit <= recordcount {
		s.ChannelMessageSend(m.ChannelID, "Sorry, you have reached your current recruitment ad limit of "+strconv.Itoa(recordlimit)+". If you would like to create additional recruitment advertisements for multiple organizations, please contact a discord staff member for assistance.")
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

	if strings.Contains(m.Content, "@everyone") || strings.Contains(m.Content, "@here") || len(m.Mentions) > 0 {
		s.ChannelMessageSend(m.ChannelID, "You are not allowed to use this org name, please remove @mentions and try again.")
		return
	}

	if len(m.Mentions) > 0 {
		s.ChannelMessageSend(m.ChannelID, "You are not allowed to use this org name, please remove any username mentions and try again.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "You have selected: **"+m.Content+"** , is this correct?")

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

	if strings.Contains(m.Content, "@everyone") || strings.Contains(m.Content, "@here") {
		s.ChannelMessageSend(m.ChannelID, "You are not allowed to use this description, please remove @everyone or @here mentions and try again.")
		return
	}

	output := "The following description was provided **(please confirm with Y/N)**: \n" + m.Content
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
		output := "**" + splitPayload[0] + "**" + "\n\n"
		output = output + splitPayload[1]

		s.ChannelMessageSend(m.ChannelID, "Your recruitment ad will be displayed as follows **(please confirm with Y/N)**\n")
		//time.Sleep(time.Second*1)
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
		output := "**" + splitPayload[0] + "**" + "\n"
		output = output + splitPayload[1]

		recruitmentAd := RecruitmentRecord{ID: uuid, OwnerID: m.Author.ID, OrgName: splitPayload[0], Description: splitPayload[1], Created: time.Now(), LastValidated: time.Now()}

		err = h.recruitmentdb.AddRecruitmentRecordToDB(recruitmentAd)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "There was an error processing your request, please contact an admin!")
			return
		}

		h.NotifyForApprovals(uuid, s, m)
		s.ChannelMessageSend(m.ChannelID, "Your advertisement has been created and submitted for moderator approval. You will received "+
			"a notification upon approval.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Recruiter Registration Cancelled.")
	return
}

func (h *RecruitmentHandler) UserIsRecruiter(userID string, s *discordgo.Session) bool {

	member, err := s.GuildMember(h.conf.DiscordConfig.GuildID, userID)
	if err != nil {
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

func (h *RecruitmentHandler) RecordExistsForUser(userID string) bool {
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

	if !h.RecordExistsForUser(m.Author.ID) {
		s.ChannelMessageSend(m.ChannelID, "Sorry, no recruitment record was found for your user ID. If you believe this is an error, please contact an admin!")
		return
	}

	records, err := h.GetRecordsForUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	list := "\n```\n"
	for i, record := range records {
		list = list + strconv.Itoa(i+1) + ") " + record.OrgName + "\n" //+ " - " + record.ID + "\n"
	}
	list = list + "\n```\n"

	s.ChannelMessageSend(m.ChannelID, "Please select an advertisement to delete: "+list)

	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	h.callback.Watch(h.SelectDeleteRecruitmentAd, uuid, "", s, m)
	return

}

func (h *RecruitmentHandler) SelectDeleteRecruitmentAd(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {
	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruiter ad deltion cancelled.")
		return
	}

	option, err := strconv.Atoi(m.Content)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid selection, ad deletion canceled.")
		return
	}
	option = option - 1

	count, err := h.GetUserRecordCount(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}
	if option < 0 || option > count-1 {
		s.ChannelMessageSend(m.ChannelID, "Invalid selection, ad deletion canceled.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Are you sure you would like to delete your recruitment ad? **(Y/N)**")
	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
		return
	}
	h.callback.Watch(h.ConfirmDeleteRecruitmentAd, uuid, strconv.Itoa(option), s, m)
	return
}

func (h *RecruitmentHandler) ConfirmDeleteRecruitmentAd(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Recruitment ad deletion cancelled.")
		return
	}

	m.Content = strings.ToLower(m.Content)
	if m.Content == "y" || m.Content == "yes" {
		records, err := h.GetRecordsForUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		option, err := strconv.Atoi(payload)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		err = h.recruitmentdb.RemoveRecruitmentRecordFromDB(records[option])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Recruitment ad deleted successfully.")
		h.NotifyDeleted(records[option], m.Author.ID, s, m)
		return
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
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if !user.Moderator {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Command 'admindelete' expects an argument (the ID of the record to be removed)")
		return
	}
	err = h.recruitmentdb.RemoveRecruitmentRecordFromDBByID(payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error removing record: "+err.Error())
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
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if !user.Moderator {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "info requires an argument!")
		return
	}

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	output := "Record info:\n```\n"
	output = output + "ID: " + record.ID + "\n"

	userrecord, err := s.User(record.OwnerID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving user: "+err.Error())
		return
	}
	output = output + "Owner: " + userrecord.Username + "(" + record.OwnerID + ")\n"
	output = output + "Approved: " + strconv.FormatBool(record.Approved) + "\n"

	approver, err := s.User(record.Approver)
	approverName := ""
	approverID := ""
	if err == nil {
		approverName = approver.Username
		approverID = approver.ID
	}
	output = output + "Approver: " + approverName + "(" + approverID + ")\n"

	loc, _ := time.LoadLocation("America/Chicago")
	output = output + "Last Run: " + record.LastRun.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"
	output = output + "Last Validated: " + record.LastValidated.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"

	expirationTime, err := h.configdb.GetValue("recruitment-expiration")
	if err != nil {
		expirationTime = int(h.conf.Recruitment.RecruitmentExpiration)
	}

	expiryMinutes := time.Duration(time.Duration(expirationTime) * time.Minute)
	expiryDate := record.LastValidated.Add(expiryMinutes)

	output = output + "Expires: " + expiryDate.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"
	output = output + "Validation Reminder Sent: " + strconv.FormatBool(record.ValidationReminderSent) + "\n"

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
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if !user.Moderator {
		return
	}

	records, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	var orgs string
	for _, record := range records {
		if record.OwnerID == userID {
			orgs = orgs + record.OrgName + ", "
		}
	}

	output := "Record info:\n```\n"

	userrecord, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving discord user: "+err.Error())
		return
	}
	output = output + "User: " + userrecord.Username + "\n"
	output = output + "User ID: " + userID + "\n"

	err = db.One("ID", userID, &user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	output = output + "Record Limit: " + strconv.Itoa(user.RecruitmentLimit) + "\n"
	output = output + "Record Count: " + strconv.Itoa(user.RecruitmentCount) + "\n"
	output = output + "Org Names: " + orgs + "\n"
	output = output + "\n```\n"
	s.ChannelMessageSend(m.ChannelID, output)

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

	if !user.Moderator {
		return
	}

	userIDSearch := ""
	if len(m.Mentions) > 0 {
		userIDSearch = m.Mentions[0].ID
	}

	page := 1
	if userIDSearch == "" {
		if len(payload) > 0 {
			page, err = strconv.Atoi(payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid page value selected - "+payload[0])
				return
			}
		}
	} else {
		if len(payload) > 1 {
			page, err = strconv.Atoi(payload[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid page value selected - "+payload[1])
				return
			}
		}
	}

	recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieiving record list!")
		return
	}

	count := 0
	for _, record := range recordlist {
		if userIDSearch != "" {
			if userIDSearch == record.OwnerID {
				count = count + 1
			}
		} else {
			count = count + 1
		}
	}
	if count == 0 {
		s.ChannelMessageSend(m.ChannelID, "No records found in search!")
		return
	}

	pagesF := float64(count) / float64(5.0)
	pages := int(math.Ceil(pagesF))

	if page > pages {
		page = pages
	}

	output := ":satellite: Recruitment records "
	output = output + "(Page " + strconv.Itoa(page) + " of " + strconv.Itoa(pages) + ")"
	output = output + ": \n```\n"

	recordCount := 0
	for _, record := range recordlist {
		if userIDSearch != "" {
			if userIDSearch == record.OwnerID {
				if recordCount < (page*5) && recordCount >= ((page-1)*5) {
					userrecord, err := s.User(record.OwnerID)
					if err == nil {
						output = output + "Org Name: " + record.OrgName + "\nOwner: " + userrecord.Username + "\nID: " + record.ID + "\n\n"
					}
				}
				recordCount = recordCount + 1
			}
		} else {
			if recordCount < (page*5) && recordCount >= ((page-1)*5) {
				userrecord, err := s.User(record.OwnerID)
				if err == nil {
					output = output + "Org Name: " + record.OrgName + "\nOwner: " + userrecord.Username + "\nID: " + record.ID + "\n\n"
				}
			}
			recordCount = recordCount + 1
		}
	}
	output = output + "\n```\n"
	//output = output + "Total Records: " + strconv.Itoa(len(recordlist))
	/*messageID, err := s.ChannelMessageSend(m.ChannelID, output)
	if err != nil {
		return
	}
	*/

	var reactions []string

	reactions = append(reactions, "⬅")
	reactions = append(reactions, "➡")

	args := strconv.Itoa(page) + "||" + userIDSearch
	err = h.reactions.Create(h.HandlePageReaction, reactions, m.ChannelID, output, args, s)

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

	if !user.Moderator {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Debug requires an argument!")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Command under construction.")
	return
}

func (h *RecruitmentHandler) ForcePost(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Moderator {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command!")
		return
	}

	h.timeoutchan <- true
	s.ChannelMessageSend(m.ChannelID, "Successfully forced the latest post from the recruitment queue")
	return
}

func (h *RecruitmentHandler) NotifyForApprovals(recordID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	unformattedChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
	if err != nil {
		return
	}

	formattedChannelID := CleanChannel(unformattedChannel)

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(recordID)
	if err != nil {
		return
	}

	owner, err := s.User(record.OwnerID)
	if err != nil {
		return
	}

	output := "ID: " + record.ID + "\n"
	output = output + "Org Name: " + record.OrgName + "\n"
	output = output + "Description: " + record.Description + "\n"

	var reactions []string
	reactions = append(reactions, "✅") // :white_check_mark:
	reactions = append(reactions, "❎") // :negative_squared_cross_mark:

	packedarg := ":rotating_light: An advertisement has been created for \"**" + record.OrgName + "\"** by " +
		"" + owner.Username + " and requires moderation:\n```\n" + output + "\n```\n"

	err = h.reactions.Create(h.HandlePendingCreatedReaction, reactions, formattedChannelID, packedarg, record.ID, s)
	if err != nil {
		fmt.Println(err.Error())
	}
	//s.ChannelMessageSend(formattedChannelID, )
	return
}

func (h *RecruitmentHandler) HandlePageReaction(reaction string, payload string, s *discordgo.Session, m interface{}) {

	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	splitstring := strings.Split(payload, "||")
	if len(splitstring) != 2 {
		return // Error
	}
	pagestring := splitstring[0]
	userIDSearch := splitstring[1]
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()

	recordlist, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error retrieving record list!")
		return
	}

	count := 0
	if userIDSearch != "" {
		for _, record := range recordlist {
			if userIDSearch == record.OwnerID {
				count = count + 1
			}
		}
	} else {
		count = len(recordlist)
	}

	if count == 0 {
		s.ChannelMessageSend(channelID, "No records found in search!")
		return
	}

	page, err := strconv.Atoi(pagestring)
	if err != nil {
		page = 1
	}

	pagesF := float64(count) / float64(5.0)
	pages := int(math.Ceil(pagesF))

	if reaction == "➡" {
		page = page + 1
		if page > pages {
			page = pages
		}
	}
	if reaction == "⬅" {
		page = page - 1
		if page == 0 {
			return
		}
	}

	output := ":satellite: Recruitment records "
	output = output + "(Page " + strconv.Itoa(page) + " of " + strconv.Itoa(pages) + ")"
	output = output + ": \n```\n"

	recordCount := 0
	for _, record := range recordlist {
		if userIDSearch != "" {
			if userIDSearch == record.OwnerID {
				if recordCount < (page*5) && recordCount >= ((page-1)*5) {
					userrecord, err := s.User(record.OwnerID)
					if err == nil {
						output = output + "Org Name: " + record.OrgName + "\nOwner: " + userrecord.Username + "\nID: " + record.ID + "\n\n"
					}
				}
				recordCount = recordCount + 1
			}
		} else {
			if recordCount < (page*5) && recordCount >= ((page-1)*5) {
				userrecord, err := s.User(record.OwnerID)
				if err == nil {
					output = output + "Org Name: " + record.OrgName + "\nOwner: " + userrecord.Username + "\nID: " + record.ID + "\n\n"
				}
			}
			recordCount = recordCount + 1
		}
	}
	output = output + "\n```\n"

	sentmsg, err := s.ChannelMessageEdit(channelID, messageID, output)
	if err != nil {
		fmt.Println(sentmsg.ID + " " + err.Error())
	}

	h.reactions.UnWatch(channelID, messageID)
	args := strconv.Itoa(page) + "||" + userIDSearch
	h.reactions.Watch(h.HandlePageReaction, messageID, channelID, args, s)
	return
}

func (h *RecruitmentHandler) HandlePendingCreatedReaction(reaction string, recordID string, s *discordgo.Session, m interface{}) {

	//fmt.Println(reaction + " - " + recordID)
	/*
		if reflect.TypeOf(m) == reflect.TypeOf(discordgo.MessageReactionAdd{}) {
			fmt.Println("Add found")
			//h.reactions.UnWatch(reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String(),
			//	reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String())
		}

		if reflect.TypeOf(m) == reflect.TypeOf(discordgo.MessageReactionRemove{}) {
			fmt.Println("Remove found")
			//h.reactions.UnWatch(reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String(),
			//	reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String())
		}
	*/

	unformattedChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
	if err != nil {
		return
	}
	formattedChannel := CleanChannel(unformattedChannel)

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(recordID)
	if err != nil {
		s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
		return
	}

	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	user, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
		return
	}

	username := user.Username
	botuser, err := h.userdb.GetUser(userID)
	if err != nil {
		s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
		return
	}
	if !botuser.Moderator {
		s.ChannelMessageSend(formattedChannel, "Error: "+user.Mention()+" is not a bot moderator!")
		return
	}

	if reaction == "✅" {
		record.Approved = true
		err = h.recruitmentdb.UpdateRecruitmentRecord(record)
		if err != nil {
			s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
			return
		}

		h.NotifyOwner(record.ID, true, "", s)

		s.ChannelMessageSend(formattedChannel, "Advertisement for \"**"+record.OrgName+"**\" has been approved by "+username)

		h.reactions.UnWatch(reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String(),
			reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String())
		return
	}
	if reaction == "❎" {
		record.Approved = false

		err = h.recruitmentdb.UpdateRecruitmentRecord(record)
		if err != nil {
			s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
			return
		}

		h.NotifyOwner(record.ID, false, "Please contact "+username+" for more information.", s)

		h.recruitmentdb.RemoveRecruitmentRecordFromDBByID(record.ID)

		s.ChannelMessageSend(formattedChannel, "Advertisement for \"**"+record.OrgName+"**\" has been rejected by "+
			""+username+" and removed from the database.")

		h.reactions.UnWatch(reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String(),
			reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String())
		return
	}

}

func (h *RecruitmentHandler) NotifyDeleted(record RecruitmentRecord, userID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	unformattedChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
	if err != nil {
		return
	}

	user, err := s.User(userID)
	if err != nil {
		return
	}

	formattedChannelID := CleanChannel(unformattedChannel)
	s.ChannelMessageSend(formattedChannelID, "Record for **"+record.OrgName+"** has been deleted by "+user.Username+".")
	return
}

func (h *RecruitmentHandler) ModifyApproval(command string, payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Moderator {
		s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command!")
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, command+" requires an argument!")
		return
	}

	recordID := payload[0]

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(recordID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	approved := false
	if command == "approve" {
		approved = true
	} else if command == "deny" || command == "reject" {

		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "An advertisement application denial requires a reason. If you do not wish to provide a reason to the creator, you can use the word 'shadow' to skip sending them a rejection notification.")
			return
		}

		approved = false
	} else {
		s.ChannelMessageSend(m.ChannelID, "Unknown error!")
		return
	}

	// If we aren't changing anything just skip
	if record.Approved {
		s.ChannelMessageSend(m.ChannelID, "Record already approved, doing nothing.")
		return
	}

	record.Approved = approved

	if approved {
		record.ApprovedDate = time.Now()
		record.Approver = m.Author.ID

		err = h.recruitmentdb.UpdateRecruitmentRecord(record)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		user, err := s.User(m.Author.ID)
		username := ""
		if err == nil {
			username = user.Username
		}
		s.ChannelMessageSend(m.ChannelID, "Advertisement for \"**"+record.OrgName+"**\" has been approved by "+username)
		h.NotifyOwner(record.ID, approved, "", s)
		return
	} else {
		if payload[1] == "shadow" {
			h.recruitmentdb.RemoveRecruitmentRecordFromDBByID(record.ID)
			return
		}
		reason := ""
		for i, word := range payload {
			if i == 1 {
				reason = word
			}
			if i > 1 {
				reason = reason + " " + word
			}
		}

		h.NotifyOwner(record.ID, approved, reason, s)
		h.recruitmentdb.RemoveRecruitmentRecordFromDBByID(record.ID)

		user, err := s.User(m.Author.ID)
		username := ""
		if err == nil {
			username = user.Username
		}
		s.ChannelMessageSend(m.ChannelID, "Advertisement for \"**"+record.OrgName+"**\" has been rejected by "+
			""+username+" and removed from the database.\n```\n"+reason+"\n```\n")
		return
	}

}

func (h *RecruitmentHandler) NotifyOwner(recordID string, approved bool, reason string, s *discordgo.Session) {

	record, err := h.recruitmentdb.GetRecruitmentRecordFromDB(recordID)
	if err != nil {
		unformattedChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
		if err != nil {
			return
		}
		formattedChannel := CleanChannel(unformattedChannel)
		s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
		return
	}

	userprivatechannel, err := s.UserChannelCreate(record.OwnerID)
	if err != nil {
		unformattedChannel, err := h.configdb.GetSetting("recruitment-admin-channel")
		if err != nil {
			return
		}
		formattedChannel := CleanChannel(unformattedChannel)
		s.ChannelMessageSend(formattedChannel, "Error: "+err.Error())
		return
	}

	if approved {
		s.ChannelMessageSend(userprivatechannel.ID, "Your advertisement has been approved")
		return
	} else {
		s.ChannelMessageSend(userprivatechannel.ID, "Your advertisement has been rejected: \n```"+reason+"\n```\n")
		return
	}

}

func (h *RecruitmentHandler) QueueInfo(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	/*
		// Grab our sender ID to verify if this user has permission to use this command
		db := h.db.rawdb.From("Users")
		var user User
		err := db.One("ID", m.Author.ID, &user)
		if err != nil {
			fmt.Println("error retrieving user:" + m.Author.ID)
			return
		}

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command!")
			return
		}
	*/

	displayRecordDB, err := h.recruitmentdb.GetAllRecruitmentDisplayDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	queuelen := len(displayRecordDB)

	output := ":satellite: Current Recruitment Advertisement Queue: ```\n"
	output = output + "Records in Queue: " + strconv.Itoa(queuelen) + "\n"
	loc, _ := time.LoadLocation("America/Chicago")
	output = output + "Last Post: " + h.lastpost.In(loc).Format("Mon Jan _2 03:04 MST 2006") + "\n"

	timercount, err := h.configdb.GetValue("recruitment-timer")
	if err != nil {
		timercount = int(h.conf.Recruitment.RecruitmentTimeout)
	}
	if timercount == 0 {
		timercount = 1
	}

	output = output + "Queue Timer: " + strconv.Itoa(timercount) + " minutes\n"

	end := h.lastpost.Add(time.Duration(int64(timercount*queuelen)) * time.Minute)
	output = output + "Estimated Time to Completion: " + end.Sub(time.Now()).Round(1*time.Second).String() + "\n"
	output = output + "Queue Shuffle Chance: " + strconv.Itoa(h.conf.Recruitment.RecruitmentShuffleCount) + "\n"

	output = output + "\nPending List: "

	pending := ""
	for i, record := range displayRecordDB {
		recruitmentRecord, err := h.recruitmentdb.GetRecruitmentRecordFromDB(record.RecruitmentID)
		if err == nil {
			if i == len(displayRecordDB)-1 {
				pending = pending + recruitmentRecord.OrgName
			} else {
				pending = pending + recruitmentRecord.OrgName + ", "
			}
		}
	}
	output = output + pending + "\n\n"

	output = output + "\n```\n"

	//fmt.Println(output)

	s.ChannelMessageSend(m.ChannelID, output)
	return
}

func (h *RecruitmentHandler) GetUserRecordLimit(userid string) (limit int, err error) {
	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err = db.One("ID", userid, &user)
	if err != nil {
		return 0, err
	}
	return user.RecruitmentLimit, nil
}

func (h *RecruitmentHandler) SetUserRecordLimit(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	// Grab our sender ID to verify if this user has permission to use this command
	db := h.db.rawdb.From("Users")
	var user User
	err = db.One("ID", m.Author.ID, &user)
	if err != nil {
		fmt.Println("error retrieving user:" + m.Author.ID)
		return
	}

	if !user.Moderator {
		return
	}

	if len(payload) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Error: setlimit requires a user mention and limit argument")
		return
	}
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Error: setlimit requires a user mention and limit argument")
		return
	}

	limit, err := strconv.Atoi(payload[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	err = db.One("ID", m.Mentions[0].ID, &user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	user.RecruitmentLimit = limit
	err = h.userdb.UpdateUserRecord(user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "User record limit updated to "+payload[1])
	return
}

func (h *RecruitmentHandler) GetUserRecordCount(userid string) (count int, err error) {

	recruitmentrecords, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		return 0, err
	}

	for _, record := range recruitmentrecords {
		if record.OwnerID == userid {
			count = count + 1
		}
	}
	return count, nil
}

func (h *RecruitmentHandler) GetRecordsForUser(userid string) (records []RecruitmentRecord, err error) {

	recruitmentrecords, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		return records, err
	}

	for _, record := range recruitmentrecords {
		if record.OwnerID == userid {
			records = append(records, record)
		}
	}
	return records, nil
}

func (h *RecruitmentHandler) FixUsers() (err error) {
	db := h.db.rawdb.From("Users")
	var users []User
	err = db.All(&users)
	if err != nil {
		return err
	}

	for _, user := range users {
		user.RecruitmentLimit = 1
		user.RecruitmentCount = 0
		h.userdb.UpdateUserRecord(user)
	}

	recruitmentrecords, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		return err
	}

	for _, record := range recruitmentrecords {
		var user User
		err = db.One("ID", record.OwnerID, &user)
		if err != nil {
			return err
		}

		user.RecruitmentCount = user.RecruitmentCount + 1
		h.userdb.UpdateUserRecord(user)
	}

	return nil
}

func (h *RecruitmentHandler) FixAds() (err error) {
	recruitmentrecords, err := h.recruitmentdb.GetAllRecruitmentDB()
	if err != nil {
		return err
	}

	for _, record := range recruitmentrecords {
		record.Approved = true
		record.Created = time.Now()
		record.ApprovedDate = time.Now()
		record.LastValidated = time.Now()
		h.recruitmentdb.UpdateRecruitmentRecord(record)
	}

	return nil
}

// ShuffleRecords function
// We use this for shuffling our record list every iteration so we don't lose records on a bot restart
func (h *RecruitmentHandler) ShuffleRecords(DisplayRecords []RecruitmentDisplayRecord) (ShuffledRecords []RecruitmentDisplayRecord) {

	count := rand.Intn(h.conf.Recruitment.RecruitmentShuffleCount)

	for pass := 0; pass < count; pass++ {
		for i := len(DisplayRecords)/2 - 1; i >= 0; i-- {
			opp := len(DisplayRecords) - 1 - i
			DisplayRecords[i], DisplayRecords[opp] = DisplayRecords[opp], DisplayRecords[i]
		}

		for i := len(DisplayRecords) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			DisplayRecords[i], DisplayRecords[j] = DisplayRecords[j], DisplayRecords[i]
		}

		for i := len(DisplayRecords)/2 - 1; i >= 0; i-- {
			opp := len(DisplayRecords) - 1 - i
			DisplayRecords[i], DisplayRecords[opp] = DisplayRecords[opp], DisplayRecords[i]
		}
	}

	return DisplayRecords
}
