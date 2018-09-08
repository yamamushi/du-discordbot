package main

import (
	"github.com/bwmarrin/discordgo"
	"gopkg.in/mgo.v2"
	"strings"
	"time"
)

type BackerHandler struct {
	db              *DBHandler
	callback        *CallbackHandler
	conf            *Config
	backerInterface *BackerInterface
}

func (h *BackerHandler) Init() {

	h.backerInterface = &BackerInterface{db: h.db, conf: h.conf}

}

func (h *BackerHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !SafeInput(s, m, h.conf) {
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if command == "migrateauth" {
		if !user.Owner {
			return
		}

		s.ChannelMessageSend(m.ChannelID, "DB Migration Started - This may take a while!")
		time.Sleep(5*time.Second)
		err = h.MigrateDB()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		//s.ChannelMessageSend(m.ChannelID, "DB Migration Successful")
		s.ChannelMessageSend(m.ChannelID, "DB Migration [Under Construction]")
		return
	}

	if command == "backerauth" || command == "atvauth" || command == "forumauth" {

		if len(payload) > 0 {
			if strings.ToLower(payload[0]) == "help"{
				s.ChannelMessageSend(m.ChannelID, ":bulb: Forum Auth Tutorial - https://www.youtube.com/watch?v=tPZuxhz6KeE")
				return
			}
		}

		userprivatechannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error initializing backerauth.")
			return
		}

		hashedid := h.backerInterface.HashUserID(m.Author.ID)
		output := ":satellite:**Authorization**:satellite: " +
			"\nIn order to validate your access to the pre-alpha sections of this discord, you must first validate " +
			"your backer status or ATV status through the Dual Universe forum.\n\n"
		output += ":one: To complete this process, please post the following text on your " +
			"public message feed through your **forum profile**:"
		output += "\n```"
		output += "discordauth:" + hashedid + "```"
		output += ":bulb: If you do not see the public message feed on your profile, you need to enable status " +
			"updates in your forum account settings.\n It's the first option in Basic Info at the top of the " +
			"**edit profile** settings window. You can disable it after this registration process is complete."

		output += "\n:bulb: If your forum account needs moderator approval, post an introduction message in the Arkship Pub " +
			"subforum so everyone knows you're a real person."

			output += "\n\n:two: Once you have posted your discordauth key, please reply to this message with the " +
			"following **command** to complete the validation process:\n"
		output += "```"
		output += "~linkprofile <url of your forum profile>"
		output += "\n```\n"
		output += "If you continue to have issues with this process, please contact a discord moderator for assistance.\n" +
			"(This message may not display properly on mobile due to a discord bug!)"

		s.ChannelMessageSend(userprivatechannel.ID, output)
		return
	}
	if command == "linkprofile" {

		userprivatechannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error initializing backerauth.")
			return
		}

		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Error: <forumauth> requires an argument!")
			return
		}

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{h.conf.DBConfig.MongoHost},
			Timeout:  30 * time.Second,
			Database: h.conf.DBConfig.MongoDB,
			Username: h.conf.DBConfig.MongoUser,
			Password: h.conf.DBConfig.MongoPass,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

		if h.backerInterface.UserValidated(m.Author.ID, *c) {
			s.ChannelMessageSend(m.ChannelID, "Error: user already validated, contact a discord admin!")
			return
		}

		profileurl := payload[0]
		if string(profileurl[len(profileurl)-1]) != "/" {
			profileurl = profileurl + "/"
		}

		err = h.backerInterface.ForumAuth(profileurl, m.Author.ID, *c)
		if err != nil {
			output := "Error validating account: " + err.Error()
			s.ChannelMessageSend(userprivatechannel.ID, output)
			return
		}

		if h.backerInterface.UserValidated(m.Author.ID, *c) {

			err := h.UpdateRoles(s, m, m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(userprivatechannel.ID, "Could not update user roles: "+err.Error()+" , please contact a discord administrator")
				return
			}

			output := "User account validated, your discord roles will be adjusted accordingly."
			s.ChannelMessageSend(userprivatechannel.ID, output)
			return
		}

		s.ChannelMessageSend(userprivatechannel.ID, "Could not validate account.")
		return
	}
	if command == "resetauth" {
		if !user.Admin {
			// Don't even bother responding, just silently fail
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command+" expects a user mention.")
			return
		}
		if len(m.Mentions) > 1 {
			s.ChannelMessageSend(m.ChannelID, command+" too many users selected.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Resetting Forum Account for : "+m.Mentions[0].Username+" Confirm? (Y/N)")

		message := m.Mentions[0].Username + "||" + m.Mentions[0].ID
		uuid, err := GetUUID()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Fatal Error generating UUID: "+err.Error())
			return
		}
		h.callback.Watch(h.ResetBackerConfirm, uuid, message, s, m)
		return
	}
	if command == "forumrole" || command == "rerunforumrole" || command == "runroles" {
		if !user.Admin {
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command+" expects a user mention.")
			return
		}
		if len(m.Mentions) > 1 {
			s.ChannelMessageSend(m.ChannelID, command+" too many users selected.")
			return
		}

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{h.conf.DBConfig.MongoHost},
			Timeout:  30 * time.Second,
			Database: h.conf.DBConfig.MongoDB,
			Username: h.conf.DBConfig.MongoUser,
			Password: h.conf.DBConfig.MongoPass,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

		if !h.backerInterface.UserValidated(m.Mentions[0].ID, *c) {
			s.ChannelMessageSend(m.ChannelID, "Selected user has not linked their profile yet.")
			return
		}

		err = h.UpdateRoles(s, m, m.Mentions[0].ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not update user roles: "+err.Error()+" , please contact a discord administrator")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Roles for "+m.Mentions[0].Mention()+" updated.")
		return
	}
	if command == "adminlink" {
		if !user.Admin{
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command+" expects a user mention.")
			return
		}
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Error: <adminlink> requires an argument!")
			return
		}

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{h.conf.DBConfig.MongoHost},
			Timeout:  30 * time.Second,
			Database: h.conf.DBConfig.MongoDB,
			Username: h.conf.DBConfig.MongoUser,
			Password: h.conf.DBConfig.MongoPass,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

		if h.backerInterface.UserValidated(m.Mentions[0].ID, *c) {
			h.ResetAuth(m.Mentions[0].ID, s, m)
		}

		profileurl := payload[1]
		if string(profileurl[len(profileurl)-1]) != "/" {
			profileurl = profileurl + "/"
		}

		err = h.backerInterface.ForumAuth(profileurl, m.Mentions[0].ID, *c)
		if err != nil {
			output := "Error validating account: " + err.Error()
			s.ChannelMessageSend(m.ChannelID, output)
			return
		}

		if h.backerInterface.UserValidated(m.Mentions[0].ID, *c) {

			err := h.UpdateRoles(s, m, m.Mentions[0].ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not update user roles: "+err.Error())
				return
			}

			output := "User account validated, discord roles will be adjusted accordingly."
			s.ChannelMessageSend(m.ChannelID, output)
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Could not validate account.")
		return
	}
	if command == "forumprofile" {
		//if !user.Admin{
		//	return
		//}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command+" expects a user mention.")
			return
		}
		if len(m.Mentions) > 1 {
			s.ChannelMessageSend(m.ChannelID, command+" too many users selected.")
			return
		}
		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{h.conf.DBConfig.MongoHost},
			Timeout:  30 * time.Second,
			Database: h.conf.DBConfig.MongoDB,
			Username: h.conf.DBConfig.MongoUser,
			Password: h.conf.DBConfig.MongoPass,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

		if !h.backerInterface.UserValidated(m.Mentions[0].ID, *c) {
			s.ChannelMessageSend(m.ChannelID, "Selected user has not linked their profile yet.")
			return
		}

		record, err := h.backerInterface.GetRecordFromDB(m.Mentions[0].ID, *c)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		output := "User record for " + m.Mentions[0].Mention() + "\n"
		output += "Profile: " + record.ForumProfile + "\n"
		output += "```"
		output += "\nFounder Status: " + record.BackerStatus
		output += "\nPreAlpha Status: " + record.PreAlpha
		output += "\nATV Status: " + record.ATV
		output += "\n```\n"
		s.ChannelMessageSend(m.ChannelID, output)
		return
	}
	if command == "forumprofilebyid" {
		if !user.Admin {
			return
		}

		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Error: forumprofilebyid expects an argument!")
			return
		}

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{h.conf.DBConfig.MongoHost},
			Timeout:  30 * time.Second,
			Database: h.conf.DBConfig.MongoDB,
			Username: h.conf.DBConfig.MongoUser,
			Password: h.conf.DBConfig.MongoPass,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

		mentionid := payload[0]

		if !h.backerInterface.UserValidated(mentionid, *c) {
			s.ChannelMessageSend(m.ChannelID, "Selected user has not linked their profile yet.")
			return
		}

		record, err := h.backerInterface.GetRecordFromDB(mentionid, *c)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		member, err := s.State.Member(h.conf.DiscordConfig.GuildID, mentionid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		usermention := member.User.Mention()

		output := "User record for " + usermention + "\n"
		output += "Profile: " + record.ForumProfile + "\n"
		output += "```"
		output += "Founder Status: " + record.BackerStatus
		if record.ATV == "true" {
			output += "\nATV Status: true"
		} else {
			output += "\nATV Status: false"
		}
		output += "\n```\n"
		s.ChannelMessageSend(m.ChannelID, output)
		return
	}
	if command == "debugroles" {
		if !user.Admin {
			return
		}

		userprivatechannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error initializing backerauth.")
			return
		}

		output := ":bulb: Backer Roles for this server: \n```"
		for _, role := range s.State.Guilds[0].Roles {
			if strings.Contains(role.Name, "Founder") || strings.Contains(role.Name, "Supporter"){
				output = output + "\n" + role.Name + " : " + role.ID
			}
		}
		output = output + "\n```\n"
		s.ChannelMessageSend(userprivatechannel.ID, output)
		return
	}
	if command == "repairbackers" {
		if !user.Owner {
			return
		}

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:    []string{h.conf.DBConfig.MongoHost},
			Timeout:  30 * time.Second,
			Database: h.conf.DBConfig.MongoDB,
			Username: h.conf.DBConfig.MongoUser,
			Password: h.conf.DBConfig.MongoPass,
		}

		session, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

		records, err := h.backerInterface.GetAllBackers(*c)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}

		for _, record := range records {
			if record.BackerStatus == "Gold Founder" || record.BackerStatus == "Sapphire Founder" || record.BackerStatus == "Ruby Founder" ||
				record.BackerStatus == "Emerald Founder" || record.BackerStatus == "Diamond Founder" || record.BackerStatus == "Kyrium Founder" ||
				record.BackerStatus == "Patron" || record.ATV == "true" {
				record.PreAlpha = "true"
				err = h.backerInterface.SaveRecordToDB(record, *c)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
					return
				}
			}
		}
		s.ChannelMessageSend(m.ChannelID, "Records repaired")
		return
	}
}

func (h *BackerHandler) MigrateDB() (err error){

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		//log.Println("Could not connect to mongo: ", err.Error())
		return err
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	backerrecords, err := h.backerInterface.GetAllBackersDeprecated()
	if err != nil {
		return err
	}

	c := session.DB("duauthbot").C(h.conf.DBConfig.BackerRecordColumn)

	for _, record := range backerrecords {
		_, err = c.UpsertId(record.UserID, record)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *BackerHandler) ResetBackerConfirm(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Reset Backer Command Cancelled")
		return
	}

	if m.Content == "Y" || m.Content == "y" {
		splitpayload := strings.Split(payload, "||")
		username := splitpayload[0]
		userid := splitpayload[1]

		err := h.ResetAuth(userid, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error resetting auth: " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: "+username+" backer status reset.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Backer Reset Cancelled")
	return
}

func (h *BackerHandler) ResetAuth(userid string, s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
		return
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)


	atvStatus, err := h.backerInterface.GetATVStatus(userid, *c)
	/*if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving ATV Status")
		return
	}*/

	prealphaStatus, err := h.backerInterface.GetPreAlphaStatus(userid, *c)
	/*if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}*/

	backerStatus, err := h.backerInterface.GetBackerStatus(userid, *c)
	/*if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}*/

	err = h.backerInterface.ResetUser(userid, *c)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not reset user: "+userid+" : "+err.Error())
		return
	}

	if backerStatus == "Iron Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.IronRoleID)

	} else if backerStatus == "Contributor" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ContributorRoleID)

	} else if backerStatus == "Bronze Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.BronzeRoleID)

	} else if backerStatus == "Sponsor" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SponsorRoleID)

	} else if backerStatus == "Silver Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SilverRoleID)

	} else if backerStatus == "Patron" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PatronRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)

	} else if backerStatus == "Gold Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.GoldRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)

	} else if backerStatus == "Sapphire Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SapphireRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)

	} else if backerStatus == "Ruby Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.RubyRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)

	} else if backerStatus == "Emerald Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.EmeraldRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)

	} else if backerStatus == "Diamond Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.DiamondRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)

	} else if backerStatus == "Kyrium Founder" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.KyriumRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
	}

	if atvStatus == "true" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ATVRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ATVForumLinkedRoleID)
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
	}

	if prealphaStatus == "true" {
		s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
	}

	s.GuildMemberRoleRemove(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ForumLinkedRoleID)
	return nil
}

func (h *BackerHandler) UpdateRoles(s *discordgo.Session, m *discordgo.MessageCreate, userid string) (err error) {

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
		return
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

	atvStatus, err := h.backerInterface.GetATVStatus(userid, *c)
	if err != nil {
		return err
	}

	prealphaStatus, err := h.backerInterface.GetPreAlphaStatus(userid, *c)
	if err != nil {
		return err
	}

	backerStatus, err := h.backerInterface.GetBackerStatus(userid, *c)
	if err != nil {
		return err
	}

	notify := false
	if backerStatus == "Iron Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.IronRoleID)
	} else if backerStatus == "Contributor" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ContributorRoleID)
	} else if backerStatus == "Bronze Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.BronzeRoleID)
	} else if backerStatus == "Sponsor" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SponsorRoleID)
	} else if backerStatus == "Silver Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SilverRoleID)
	} else if backerStatus == "Patron" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PatronRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	} else if backerStatus == "Gold Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.GoldRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	} else if backerStatus == "Sapphire Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SapphireRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	} else if backerStatus == "Ruby Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.RubyRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	} else if backerStatus == "Emerald Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.EmeraldRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	} else if backerStatus == "Diamond Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.DiamondRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	} else if backerStatus == "Kyrium Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.KyriumRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	}

	if prealphaStatus == "true" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	}

	if atvStatus == "true" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ATVRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ATVForumLinkedRoleID)
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.PreAlphaForumLinkedRole)
		notify = true
	}

	s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ForumLinkedRoleID)
	if notify {
		h.NotifyNDAChannelOnAuth(s, userid)
	}
	return nil
}


func (h *BackerHandler) NotifyNDAChannelOnAuth(s *discordgo.Session, userid string){

	user, err := s.User(userid)
	if err != nil {
		return
	}

	s.ChannelMessageSend(h.conf.RolesConfig.NDAChannelID, user.Mention() + " has been authorized as having pre-alpha " +
							"access, and can now use the NDA Discord channels. Congrats!")
	return
}
