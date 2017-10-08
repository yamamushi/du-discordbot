package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

type BackerHandler struct {

	db *DBHandler
	callback *CallbackHandler
	conf *Config
	backerInterface *BackerInterface
}


func (h *BackerHandler) Init(){

	h.backerInterface = &BackerInterface{db: h.db}

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

	if command == "backerauth" || command == "atvauth" || command == "forumauth"{

		userprivatechannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error initializing backerauth.")
			return
		}

		hashedid := h.backerInterface.HashUserID(m.Author.ID)
		output := "In order to validate your access to the pre-alpha sections of this discord, you must first validate " +
			"your backer status or ATV status through the Dual Universe forum.\n\n"
		output += "To complete this process, please submit the following text into your " +
			"public message feed through your **forum profile** (If you do not see the public message feed on your profile, " +
				"you need to enable status updates in your forum account settings. It's the first option in Basic Info at the " +
					"top of the **edit profile** settings window. You can disable it after the registration process is complete):"
		output += "\n\n**Remember to paste this as plaintext or paste and match style, or this will not work otherwise!**\n"
		output += "```discordauth:"+hashedid+"```"
		output += "\n**99% of the time if you are having issues it is because you did not use Plaintext**\n\n"
		output += "Once you have finished this step, please reply to this message with the " +
			"following **command** to complete the validation process:\n"
		output += "```"
		output += "~linkprofile <url of your forum profile>"
		output += "\n```\n"
		output += "If you continue to have issues with this process, please contact a discord moderator for assistance."

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

		if h.backerInterface.UserValidated(m.Author.ID){
			s.ChannelMessageSend(m.ChannelID, "Error: user already validated, contact a discord admin!")
			return
		}

		err = h.backerInterface.ForumAuth(payload[0], m.Author.ID)
		if err != nil {
			output := "Error validating account: " + err.Error()
			s.ChannelMessageSend(userprivatechannel.ID, output)
			return
		}

		if h.backerInterface.UserValidated(m.Author.ID) {

			err := h.UpdateRoles(s, m, m.Author.ID)
			if err != nil{
				s.ChannelMessageSend(userprivatechannel.ID, "Could not update user roles: " + err.Error() + " , please contact a discord administrator")
				return
			}

			output := "User account validated, your discord roles will be adjusted accordingly."
			s.ChannelMessageSend(userprivatechannel.ID, output)
			return
		}

		s.ChannelMessageSend(userprivatechannel.ID, "Could not validate account.")
		return
	}
	if command == "resetforum" {
		if !user.Admin {
			// Don't even bother responding, just silently fail
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command + " expects a user mention.")
			return
		}
		if len(m.Mentions) > 1 {
			s.ChannelMessageSend(m.ChannelID, command + " too many users selected.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Resetting Forum Account for : "+m.Mentions[0].Username+" Confirm? (Y/N)")

		message := m.Mentions[0].Username + " " + m.Mentions[0].ID
		h.callback.Watch(h.ResetBackerConfirm, GetUUID(), message, s, m)
		return
	}
	if command == "forumrole" || command == "rerunforumrole" || command == "runroles"{
		if !user.Admin{
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command + " expects a user mention.")
			return
		}
		if len(m.Mentions) > 1 {
			s.ChannelMessageSend(m.ChannelID, command + " too many users selected.")
			return
		}

		if !h.backerInterface.UserValidated(m.Mentions[0].ID) {
			s.ChannelMessageSend(m.ChannelID, "Selected user has not linked their profile yet.")
			return
		}

			err := h.UpdateRoles(s, m, m.Mentions[0].ID)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Could not update user roles: " + err.Error() + " , please contact a discord administrator")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Roles for " + m.Mentions[0].Mention() + " updated.")
		return
	}
	if command == "forumprofile" {
		if !user.Admin{
			return
		}
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, command + " expects a user mention.")
			return
		}
		if len(m.Mentions) > 1 {
			s.ChannelMessageSend(m.ChannelID, command + " too many users selected.")
			return
		}

		if !h.backerInterface.UserValidated(m.Mentions[0].ID) {
			s.ChannelMessageSend(m.ChannelID, "Selected user has not linked their profile yet.")
			return
		}

		record, err := h.backerInterface.GetRecordFromDB(m.Mentions[0].ID)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}

		output := "User record for " + m.Mentions[0].Mention() + "\n"
		output += "```"
		output += "Founder Status: " + record.BackerStatus
		if record.ATV == "true" {
			output += "ATV Status: true"
		} else {
			output += "ATV Status: false"
		}
		output += "Profile: " + record.ForumProfile
		output += "\n```\n"
		s.ChannelMessageSend(m.ChannelID, output)
		return
	}

}

func (h *BackerHandler) ResetBackerConfirm(payload string, s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	if strings.HasPrefix(m.Content, cp) {
		s.ChannelMessageSend(m.ChannelID, "Reset Backer Command Cancelled")
		return
	}

	if m.Content == "Y" || m.Content == "y" {
		splitpayload := strings.Fields(payload)
		username := splitpayload[0]
		userid := splitpayload[1]

		err := h.backerInterface.ResetUser(userid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Selection Confirmed: "+username+" backer status reset.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Backer Reset Cancelled")
	return
}


func (h *BackerHandler) UpdateRoles(s *discordgo.Session, m *discordgo.MessageCreate, userid string) (err error){

	atvStatus, err := h.backerInterface.GetATVStatus(userid)
	if err != nil {
		return err
	}

	backerStatus, err := h.backerInterface.GetBackerStatus(userid)
	if err != nil {
		return err
	}

	if backerStatus == "Iron Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.IronRoleID)

	} else if backerStatus == "Bronze Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.BronzeRoleID)

	} else if backerStatus == "Silver Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SilverRoleID)

	} else if backerStatus == "Gold Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.GoldRoleID)

	} else if backerStatus == "Sapphire Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.SapphireRoleID)

	} else if backerStatus == "Ruby Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.RubyRoleID)

	} else if backerStatus == "Emerald Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.EmeraldRoleID)

	} else if backerStatus == "Diamond Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.DiamondRoleID)

	} else if backerStatus == "Kyrium Founder" {
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.KyriumRoleID)
	}

	if atvStatus == "true"{
		s.GuildMemberRoleAdd(h.conf.DiscordConfig.GuildID, userid, h.conf.RolesConfig.ATVRoleID)
	}

	return nil
}