package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"errors"
)

type PermissionsHandler struct {

	db *DBHandler
	conf *Config
	dg *discordgo.Session
	callback *CallbackHandler
	user *UserHandler
	logger *Logger

}


func (h *PermissionsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate){

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	// Verify the user account exists (creates one if it doesn't exist already)
	h.user.CheckUser(m.Author.ID)

	/*
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println("Error finding user")
		return
	}
	*/

	// Command prefix
	cp := h.conf.DUBotConfig.CP

	// Command from message content
	command := strings.Fields(m.Content)
	// We use this a bit, this is the author id formatted as a mention
	//authormention := m.Author.Mention()
	//mentions := m.Mentions

	// We don't care about commands that aren't formatted for this handler
	if len(command) < 1{
		return
	}

	command[0] = strings.TrimPrefix(command[0], cp)

	// After our command string has been trimmed down, check it against the command list
	if command[0] == "set" {
		if len(command) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<set> expects an argument.")
			return
		}
	}
	if command[0] == "promote" {
		// Run our promote command function
		h.ReadPromote(command, s, m)
		return
	}

	if command[0] == "demote" {
		// Run our promote command function
		h.ReadDemote(command, s, m)
		return
	}
	return
}


// The promote command runs using our commands array to get the promotion settings
func (h *PermissionsHandler) ReadPromote(commands []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(commands) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: promote <user> <group>")
		return
	}
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "User must be mentioned")
		return
	}

	// Grab our target user id and group
	target := m.Mentions[0].ID
	group := commands[2]


	// Get the authors user object from the database
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println("Could not find user in PermissionsHandler.ReadPromote")
		return
	}

	// Check the group argument
	if group == "owner" {
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " https://www.youtube.com/watch?v=fmz-K2hLwSI ")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to owner", s)
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "This group cannot be assigned through the promote command.")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to owner", s)
			return
		}
	}
	if group == "admin" {
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to admin", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to admin || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "smoderator" {

		if !user.Admin {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to smoderator", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to smoderator || " +
					m.Mentions[0].Mention() + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "moderator" {

		if !user.SModerator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to moderator", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to moderator || " +
					target + "||" + group + "||" + err.Error(), s)

				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "editor" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to editor", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to editor || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "agora" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to agora", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to agora || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "streamer" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to streamer", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to streamer || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "recruiter" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to recruiter", s)
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to recruiter || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, group + " is not a valid group!")
		h.logger.LogPerm( m.Author.Mention() + " attempted to promote " + m.Mentions[0].Mention() +
			" to " + group + " which does not exist", s)
		return
	}
}


// Set the given role on a user, and save the changes in the database
func (h *PermissionsHandler) Promote(userid string, group string) (err error) {

	// Get user from the database using the userid
	user, err := h.user.GetUser(userid)
	if err != nil{
		return err
	}

	// Checks if a user is in a group based on the group string
	if user.CheckRole(group) {
		return errors.New("User Already in Group "+group+"!")
	}

	// Open the "Users" bucket in the database
	db := h.db.rawdb.From("Users")

	// Assign the group to our target user
	user.SetRole(group)

	// Save the user changes in the database
	db.Update(&user)
	return nil
}


// The promote command runs using our commands array to get the promotion settings
func (h *PermissionsHandler) ReadDemote(commands []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(commands) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: demote <user> <group>")
		return
	}
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "User must be mentioned")
		return
	}

	// Grab our target user id and group
	target := m.Mentions[0].ID
	group := commands[2]


	// Get the authors user object from the database
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println(err.Error())
		return
	}

	// Check the group argument
	if group == "owner" {
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " https://www.youtube.com/watch?v=7qnd-hdmgfk ")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to owner", s)
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "This group cannot be assigned through the promote command.")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to owner", s)
			return
		}
	}
	if group == "admin" {
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to admin", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to admin || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been set to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been demoted to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "smoderator" {

		if !user.Admin {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to smoderator", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to smoderator || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been set to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been demoted to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "moderator" {

		if !user.SModerator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run promote to moderator", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to moderator || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been set to the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been demoted to the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "editor" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to editor", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to editor || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been removed from the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "agora" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to agora", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to agora || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been removed from the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "streamer" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to streamer", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to streamer || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been removed from the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	}
	if group == "recruiter" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to recruiter", s)
			return
		} else {
			err = h.Demote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				h.logger.LogPerm(m.Author.Mention() + " attempted to run demote to recruiter || " +
					target + "||" + group + "||" + err.Error(), s)
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been removed from the " + group + " group.")
			h.logger.LogPerm(m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention(), s)
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, group + " is not a valid group!")
		h.logger.LogPerm( m.Author.Mention() + " attempted to demote " + m.Mentions[0].Mention() +
			" to " + group + " which does not exist", s)
		return
	}
}

// Set the given role on a user, and remove all promotions above the group
// If it is the lowest tier of group, that group is removed from the user
func (h *PermissionsHandler) Demote(userid string, group string) (err error) {

	// Open the "Users" bucket in the database
	db := h.db.rawdb.From("Users")

	// Get user from the database using the userid
	userobject := User{}
	err = db.One("ID", userid, &userobject)
	if err != nil{
		return err
	}

	if group == "smoderator" {
		userobject.Admin = false
	}
	if group == "moderator" {
		userobject.Admin = false
		userobject.SModerator = false
	}
	if group == "agora" {
		userobject.Agora = false
	}
	if group == "recruiter" {
		userobject.Recruiter = false
	}

	if group == "streamer" {
		userobject.Streamer = false
	}

	if group == "editor" {
		userobject.Editor = false
	}

	err = db.DeleteStruct(&userobject)
	if err != nil {
		return err
	}
	err = db.Save( &userobject)
	if err != nil {
		return err
	}

	return nil
}


