package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	//"fmt"
	"time"
	"encoding/json"
	"strconv"
)

// NotificationsHandler struct
type RoleHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	rolesDB  RoleDB
	user     *UserHandler
}

// Init function
func (h *RoleHandler) Init() {
	h.RegisterCommands()
	h.rolesDB = RoleDB{db: h.db}
}


// RegisterCommands function
func (h *RoleHandler) RegisterCommands() (err error) {
	h.registry.Register("roles", "Manage roles system", "list|add|edit|sync|remove|debug|flush|show")
	return nil
}

// Read function
func (h *RoleHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"roles") {
		if h.registry.CheckPermission("roles", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				//fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Admin {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

func (h *RoleHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command " + command + " expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "edit" {
		_, commandpayload := SplitPayload(payload)
		h.RoleEdit(commandpayload, s, m)
		return
	}
	if payload[0] == "list" {
		_, commandpayload := SplitPayload(payload)
		h.RoleList(commandpayload, s, m)
		return
	}
	if payload[0] == "add" {
		_, commandpayload := SplitPayload(payload)
		h.RoleAdd(commandpayload, s, m)
		return
	}
	if payload[0] == "sync" {
		_, commandpayload := SplitPayload(payload)
		h.RoleSync(commandpayload, s, m)
		return
	}
	if payload[0] == "remove" {
		_, commandpayload := SplitPayload(payload)
		h.RoleRemove(commandpayload, s, m)
		return
	}
	if payload[0] == "debug" {
		_, commandpayload := SplitPayload(payload)
		h.DebugRoles(commandpayload, s, m)
		return
	}
	if payload[0] == "flush" {
		_, commandpayload := SplitPayload(payload)
		h.FlushQueue(commandpayload, s, m)
		return
	}
	if payload[0] == "show" ||  payload[0] == "json" {
		_, commandpayload := SplitPayload(payload)
		h.RoleJson(commandpayload, s, m)
		return
	}
	if payload[0] == "whitelist" {
		if len(m.Mentions) == 0 {
			s.ChannelMessageSend(m.ChannelID, "usage: roles whitelist @user")
			return
		}
		h.WhitelistUser(m.Mentions[0].ID, s, m)
		return
	}
}


// HelpOutput function
func (h *RoleHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate){
	output := "Command usage for giveaway: \n"
	output = output + "```\n"
	output = output + "list: list auto roles and their id's\n"
	output = output + "add: add an auto role in JSON format\n"
	output = output + "edit: adjust the properties of an auto role\n"
	output = output + "sync: manually sync the order of all auto roles in discord\n"
	output = output + "remove: remove an auto role\n"
	output = output + "debug: provides debug output to console\n"
	output = output + "whitelist: disables autorole on a user mention\n"
	output = output + "flush: manually flush the role update queue - disabled\n"
	output = output + "json: display a role in json format\n"
	output = output + "```\n"
	s.ChannelMessageSend(m.ChannelID, output)
}



func (h *RoleHandler) RoleSynchronizer(s *discordgo.Session) {
	for true {
		// Only run every X minutes
		time.Sleep(h.conf.RolesConfig.RoleTimer * time.Minute)
		//fmt.Println("Running Synchronizer")
		memberlist, err := h.GetAllUsers(s)
		if err == nil {

			roleslist, err := h.rolesDB.GetAllRolesDB()
			if err == nil {

				for _, member := range memberlist {

					memberjoined, err := time.Parse(time.RFC3339, member.JoinedAt)
					if err == nil && !member.User.Bot {

						memberAge := time.Since(memberjoined)
						for _, role := range roleslist {

							if memberAge > role.TimeoutDuration {
								userobject, err := h.user.GetUser(member.User.ID)

								if err == nil {
									// Skip disabled autorole users
									if !userobject.DisableAutoRole {

										time.Sleep(1 * time.Second)

										roleID, _ := getRoleIDByName(s, h.conf.DiscordConfig.GuildID, role.Name)
										userHasRole := false

										for _, currentRole := range member.Roles {
											//fmt.Println("Roles: " + currentRole + " - " + roleID )
											if currentRole == roleID {
												userHasRole = true
											}
										}

										userHasAutoRole := false
										if role.ID == userobject.CurrentAutoRoleID {
											userHasAutoRole = true
										}

										// If our last updated timeout is less than the role timeout
										if userobject.LatestRoleTimeout <= role.TimeoutDuration {
											if !userHasRole && !userHasAutoRole{
												//fmt.Print("Adding: ")
												//fmt.Println(role.Name+": " + userobject.ID + " - " + userobject.LatestRoleTimeout.String() + " - " + role.TimeoutDuration.String() + " - " + roleID + " - CurrentAutoRole: " + userobject.CurrentAutoRoleID)

												// Add role to member
												uuid, err := GetUUID()
												if err == nil {
													queued := RoleQueued{ID: uuid, Remove: false, UserID: member.User.ID, RoleID: role.ID}
													h.rolesDB.AddRoleQueuedToDB(queued)

													userobject.CurrentAutoRoleID = role.ID
													userobject.LatestRoleTimeout = role.TimeoutDuration
													h.user.UpdateUserRecord(userobject)
													}
												}
										} else {
											if userHasRole {
												//fmt.Print("Removing: ")
												//fmt.Println(role.Name+": " + userobject.ID + " - " + userobject.LatestRoleTimeout.String() + " - " + role.TimeoutDuration.String() + " - " + roleID)

												// Add role to member
												uuid, err := GetUUID()
												if err == nil {
													queued := RoleQueued{ID: uuid, Remove: true, UserID: member.User.ID, RoleID: role.ID}
													h.rolesDB.AddRoleQueuedToDB(queued)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}



func (h *RoleHandler) RoleUpdater(s *discordgo.Session) {
	for true {
		time.Sleep(h.conf.RolesConfig.RoleUpdaterTimer * time.Minute)

		queuedRoles, err := h.rolesDB.GetAllRoleQueuedDB()
		if err == nil {
			//fmt.Println("\nLength of Role Updater queue - " + strconv.Itoa(len(queuedRoles)))
			for _, queuedRole := range queuedRoles {
				//fmt.Println("Parsing role: " + queuedRole.RoleID + " - " + queuedRole.UserID )
				time.Sleep(time.Second * 2)
				if queuedRole.Remove {
					role, err := h.rolesDB.GetRoleFromDB(queuedRole.RoleID)
					if err == nil {

						guildroles, _ := s.GuildRoles(h.conf.DiscordConfig.GuildID)

						for _, guildrole := range guildroles {
							if guildrole.Name == role.Name {
								//fmt.Println("Role removed: " + queuedRole.UserID + " - " + guildrole.ID)
								err = s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, queuedRole.UserID, guildrole.ID)
							}
						}


					} else {
						//fmt.Println("Error Getting role from DB: " + err.Error())
					}
				} else {
					role, err := h.rolesDB.GetRoleFromDB(queuedRole.RoleID)
					if err == nil {
						guildroles, _ := s.GuildRoles(h.conf.DiscordConfig.GuildID)

						for _, guildrole := range guildroles {
							if guildrole.Name == role.Name {
								userobject, err := h.user.GetUser(queuedRole.UserID)
								if err == nil {
									//fmt.Println("Role added: " + queuedRole.UserID + " - " + guildrole.ID)
									err = s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, queuedRole.UserID, guildrole.ID)
									h.user.UpdateUserRecord(userobject)
								} else {
									//fmt.Println("Error getting RoleID: " +err.Error())
								}
							}
						}

					} else {
						//fmt.Println("Error getting role from DB: " + err.Error())
					}
				}
				err = h.rolesDB.RemoveRoleQueuedFromDB(queuedRole)
				if err != nil {
					//fmt.Println(err.Error())
				}
			}
		}
	}
}


func (h *RoleHandler) GetAllUsers(s *discordgo.Session) (userList []*discordgo.Member, err error){

	guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	if err != nil {
		return nil, err
	}
	//fmt.Println(strconv.Itoa(len(guild.Members)))
	//for _, member := range guild.Members {
	//	fmt.Println(strconv.Itoa(i) + " - " + member.User.ID)
	//	fmt.Println(member.JoinedAt)
	//	parsedTime, _ := time.Parse(time.RFC3339, member.JoinedAt)
	//	fmt.Println(parsedTime.String())
	//}
	//s.GuildMembers(h.conf.DiscordConfig.GuildID, "", 250)
	return guild.Members, nil
}




func (h *RoleHandler) RoleEdit(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'edit' expects a formatted role, see help for usage.")
		return
	}

	if strings.ToLower(payload[0]) == "help" {
		examplePayload := "{\n\t\"Name\": \"Spectator\", \n\t\"NewName\":\"OptionalNewName\",\"\n\t\"Color\": 252525,\n\t\"timeout\" : \"45m\"\n}"
		s.ChannelMessageSend(m.ChannelID, "'edit' expects a payload formatted in json. Example: ```" +examplePayload+ "\n```" )
		return
	}

	var combined string
	for count, i := range payload {
		if count != 0 && count != len(payload) - 1 {
			combined += i + " "
		}
	}

	unpacked, err := h.UnpackRole(combined)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error unpacking payload: " + err.Error())
		return
	}

	duration, _, err := ParseDuration(unpacked.Timeout)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing timeout: " + err.Error())
		return
	}
	unpacked.TimeoutDuration = duration

	roleslist, err := h.rolesDB.GetAllRolesDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading roles DB: " + err.Error())
		return
	}

	guildroles, _ := s.GuildRoles(h.conf.DiscordConfig.GuildID)

	found := false
	for _, roleinlist := range roleslist {
		if strings.ToLower(roleinlist.Name) == strings.ToLower(unpacked.Name) {
			found = true
			unpacked.ID = roleinlist.ID

			guildRoleID := ""
			for _, guildRole := range guildroles {
				if guildRole.Name == unpacked.Name {
					guildRoleID = guildRole.ID
				}
			}

			if unpacked.NewName != "" {
				roleinlist.Name = unpacked.NewName
			}

			if unpacked.Color != roleinlist.Color {
				roleinlist.Color = unpacked.Color
			}

			if unpacked.Timeout != roleinlist.Timeout {
				roleinlist.Timeout = unpacked.Timeout
				roleinlist.TimeoutDuration = unpacked.TimeoutDuration
			}



			_, err = s.GuildRoleEdit(h.conf.DiscordConfig.GuildID, guildRoleID, roleinlist.Name, roleinlist.Color, true, 0, false )
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Could not update Discord role:\n " + err.Error())
				return
			}

			err = h.rolesDB.UpdateRoleRecord(roleinlist)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Could not update database: " + err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Role successfully updated")
			return
		}
	}

	if !found {
		s.ChannelMessageSend(m.ChannelID, "Error: Role with name " + unpacked.Name + " does not exist, did you mean to use `new` instead?")
		return
	}

}



func (h *RoleHandler) RoleAdd(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'add' expects a formatted role, see help for usage.")
		return
	}

	if strings.ToLower(payload[0]) == "help" {
		examplePayload := "{\n\t\"Name\": \"Spectator\",\n\t\"Color\": 9715417,\n\t\"timeout\" : \"15m\"\n}"
		s.ChannelMessageSend(m.ChannelID, "'add' expects a payload formatted in json. Example: ```" +examplePayload+ "\n```" )
		return
	}

	var combined string
	for count, i := range payload {
		if count != 0 && count != len(payload) - 1 {
			combined += i + " "
		}
	}

	unpacked, err := h.UnpackRole(combined)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error unpacking payload: " + err.Error())
		return
	}

	duration, _, err := ParseDuration(unpacked.Timeout)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing timeout: " + err.Error())
		return
	}
	unpacked.TimeoutDuration = duration

	roleslist, err := h.rolesDB.GetAllRolesDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading roles DB: " + err.Error())
		return
	}

	for _, roleinlist := range roleslist {
		if strings.ToLower(roleinlist.Name) == strings.ToLower(unpacked.Name) {
			s.ChannelMessageSend(m.ChannelID, "Error: Role with name " + unpacked.Name + " already exists!")
			return
		}
	}

	uuid, err := GetUUID()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error generating role ID: " + err.Error())
		return
	}
	unpacked.ID = uuid

	err = h.AddRole(unpacked, s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error adding role to Guild: " + err.Error())
		return
	}
	err = h.rolesDB.AddRoleRecordToDB(unpacked)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error saving role to database: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Role successfully added")
	return
}

func (h *RoleHandler) AddRole(record RoleRecord, s *discordgo.Session) (err error) {
	createdrole, err := s.GuildRoleCreate(h.conf.DiscordConfig.GuildID)
	if err != nil {
		return err
	}

	_, err = s.GuildRoleEdit(h.conf.DiscordConfig.GuildID, createdrole.ID, record.Name, record.Color, true, 0, false)
	return err
}

func (h *RoleHandler) UnpackRole(payload string) (unpacked RoleRecord, err error) {

	payload = strings.TrimPrefix(payload, "~giveaway new ") // This all will need to be updated later, this is just
	payload = strings.TrimPrefix(payload, "\n")             // A lazy way of cleaning the command
	payload = strings.TrimPrefix(payload, "```")
	payload = strings.TrimSuffix(payload, "```")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.Trim(payload, "```")

	unmarshallcontainer := RoleRecord{}
	if err := json.Unmarshal([]byte(payload), &unmarshallcontainer); err != nil {
		return RoleRecord{}, err
	} else {
		return unmarshallcontainer, nil
	}
}

func (h *RoleHandler) DebugRoles(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	//guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	//if err != nil {
	//	return nil, err
	//}
	//fmt.Println(strconv.Itoa(len(guild.Members)))
	//for _, member := range guild.Members {
	//	fmt.Println(strconv.Itoa(i) + " - " + member.User.ID)
	//	fmt.Println(member.JoinedAt)
	//	parsedTime, _ := time.Parse(time.RFC3339, member.JoinedAt)
	//	fmt.Println(parsedTime.String())
	//}
	//s.GuildMembers(h.conf.DiscordConfig.GuildID, "", 250)

	s.ChannelMessageSend(m.ChannelID, "Check console for debug output")
	return
}

func (h *RoleHandler) RoleList(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){

	output := "AutoRole List: ```\n"
	roleslist, err := h.rolesDB.GetAllRolesDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading roles DB: " + err.Error())
		return
	}

	for i, roleinlist := range roleslist {
		output = output + strconv.Itoa(i+1) + ") " + roleinlist.Name + " - Timeout: " + roleinlist.Timeout + "\n\n"
	}
	output = output + "\n```"
	s.ChannelMessageSend(m.ChannelID, output)
	return
}

func (h *RoleHandler) RoleInit(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'init' expects an argument.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "under construction")
	return
}

func (h *RoleHandler) RoleJson(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'json' expects an argument.")
		return
	}
	output := payload[0] + " ```\n"

	roleslist, err := h.rolesDB.GetAllRolesDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading roles DB: " + err.Error())
		return
	}
	found := false
	for _, roleinlist := range roleslist {
		if payload[0] == roleinlist.Name {
			found = true
/*
			jsonoutput, err := json.Marshal(roleinlist)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error marshalling role: " + err.Error())
				return
			}
*/

			jsonoutput, err := json.MarshalIndent(roleinlist, "", "    ")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error marshalling role: " + err.Error())
				return
			}
			output = output + string(jsonoutput)
		}
	}

	if !found {
		s.ChannelMessageSend(m.ChannelID, "Role "+payload[0]+" not found")
		return
	}


	output = output + "\n```\n"

	s.ChannelMessageSend(m.ChannelID, output)
	return
}


func (h *RoleHandler) WhitelistUser(userid string, s *discordgo.Session, m *discordgo.MessageCreate){

	userobject, err := h.user.GetUser(userid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: could not validate user - " + err.Error())
		return
	}

	userobject.DisableAutoRole = !userobject.DisableAutoRole
	err = h.user.UpdateUserRecord(userobject)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: could not save user record - " + err.Error())
		return
	}
	if userobject.DisableAutoRole {
		s.ChannelMessageSend(m.ChannelID, "User autorole disabled")
	} else {
		s.ChannelMessageSend(m.ChannelID, "User autorole enabled ")
	}
	return
}

func (h *RoleHandler) FlushQueue(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	s.ChannelMessageSend(m.ChannelID, "Disabled Command - Enabled in Dev Only")
	//h.flush()
	//s.ChannelMessageSend(m.ChannelID, "Database Flushed")
	return
}

func (h *RoleHandler) flush() (err error){

	h.rolesDB.FlushDB()

	usersdb := h.user.db.rawdb.From("Users")
	var userlist []User
	usersdb.All(&userlist)
	//fmt.Println("Db size: " + strconv.Itoa(len(userlist)))

	for _, userrecord := range userlist {
		userrecord.CurrentAutoRoleID = ""
		userrecord.LatestRoleTimeout = 0
		usersdb.DeleteStruct(&userrecord)
		usersdb.Save(&userrecord)
	}

	return nil
}


func (h *RoleHandler) RoleSync(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	s.ChannelMessageSend(m.ChannelID, "Starting role order synchronization")

	sortedlist, err := h.rolesDB.GetAllRolesDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not read roles db: " + err.Error() )
		return
	}

	discordroles, err := s.GuildRoles(h.conf.DiscordConfig.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading discord roles: " + err.Error())
		return
	}
	var orderedlist []*discordgo.Role
	var originalist []*discordgo.Role

	var guestrole discordgo.Role
	for _, discordrole := range discordroles {
		for _, item := range sortedlist {
			if discordrole.Name == item.Name {
				orderedlist = append(orderedlist, discordrole)
				item.Color = discordrole.Color
				h.rolesDB.UpdateRoleRecord(item)
			} else if discordrole.Name != "Guest"{
				originalist = append(originalist, discordrole)
			}
		}
		if discordrole.Name == "Guest" {
			guestrole = *discordrole
		}
	}

	var finallist []*discordgo.Role
	for _, item := range originalist {
		finallist = append(finallist, item)
	}

	for _, item := range orderedlist {
		finallist = append(finallist, item)
	}

	finallist = append(finallist, &guestrole)

	_, err = s.GuildRoleReorder(h.conf.DiscordConfig.GuildID, finallist)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reordering roles: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Auto Roles synchronized")
	return
}

func (h *RoleHandler) RoleRemove(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command 'remove' expects an argument.")
		return
	}

	roleDiscordID, err := getRoleIDByName(s, h.conf.DiscordConfig.GuildID, payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving role from Discord API: " + err.Error() + " \n Continuing...")
		//return
	}

	rolelist, err := h.rolesDB.GetAllRolesDB()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving role list from DB: " + err.Error())
		return
	}

	for _, role := range rolelist {
		if role.Name == payload[0] {
			err = h.rolesDB.RemoveRoleRecordFromDB(role)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error removing role from DB: " + err.Error())
				return
			}
		}
	}

	err = s.GuildRoleDelete(h.conf.DiscordConfig.GuildID, roleDiscordID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error removing role from Discord: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Role removed")
	return
}