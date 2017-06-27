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

	cp := h.conf.DUBotConfig.CP
	command := strings.Fields(m.Content)
	// We use this a bit, this is the author id formatted as a mention
	//authormention := m.Author.Mention()
	//mentions := m.Mentions

	// We don't care about commands that aren't formatted for this handler
	if len(command) < 1{
		return
	}

	command[0] = strings.TrimPrefix(command[0], cp)

	if command[0] == "set" {
		if len(command) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<set> expects an argument.")
			return
		}
	}
	if command[0] == "promote" {
		h.ReadPromote(command, s, m)
		return
	}
	return
}


func (h *PermissionsHandler) ReadPromote(commands []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(commands) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: promote <user> <group>")
		return
	}
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "User must be mentioned")
		return
	}
	target := m.Mentions[0].ID
	group := commands[2]

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println("Could not find user in PermissionsHandler.ReadPromote")
		return
	}

	if group == "owner" {
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " https://www.youtube.com/watch?v=fmz-K2hLwSI ")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "This group cannot be assigned through the promote command.")
			return
		}
	}
	if group == "admin" {
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	}
	if group == "smoderator" {

		if !user.Admin {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	}
	if group == "moderator" {

		if !user.SModerator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	}
	if group == "editor" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	}
	if group == "agora" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	}
	if group == "streamer" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	}
	if group == "recruiter" {

		if !user.Moderator {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			return
		} else {
			err = h.Promote(target, group)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention() + " has been added to the " + group + " group.")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, group + " is not a valid group!")
		return
	}
}


func (h *PermissionsHandler) Promote(userid string, group string) (err error) {

	user, err := h.user.GetUser(userid)
	if err != nil{
		return err
	}

	if user.CheckRole(group) {
		return errors.New("User Already in Group "+group+"!")
	}

	db := h.db.rawdb.From("Users")
	user.SetRole(group)

	db.Update(&user)
	return nil
}


func (h *PermissionsHandler) AdminCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate)  {

	if command[0] == "set" {
		if len(command) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<set> expects an argument.")
			return
		}
	}
	if command[0] == "promote" {
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Usage: promote <user> <group>")
			return
		}
		if command[2] == "owner" {

			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " https://www.youtube.com/watch?v=fmz-K2hLwSI ")
		}
	}
}