package main

/*

Our interface to our command registry

 */


import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"os"
	"fmt"
)

type CommandHandler struct {
	callback *CallbackHandler
	conf *Config
	db *DBHandler
	perm *PermissionsHandler
	registry *CommandRegistry
	dg *discordgo.Session
	user *UserHandler
}


func (h *CommandHandler) Init(){

	// Setup our command registry interface
	h.registry = new(CommandRegistry)
	h.registry.conf = h.conf
	if h.conf.DUBotConfig.PerPageCount < 2 {
		fmt.Println("Invalid Config Parameter Setting: [du-bot]:per_page_count must be 2 or higher!")
		os.Exit(0)
	}
	h.registry.db = h.db
	h.registry.user = h.user

}




func (h *CommandHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Check for safety
	if !SafeInput(s, m){
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		return
	}
	if !user.CheckRole("admin"){
		return
	}

	command, message := CleanCommand(m.Content, h.conf)

	if command == "command" {
		h.ReadCommand(message, s, m)

	}
}


func (h *CommandHandler) ReadCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate){

	if len(message) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<command> requires an argument")
		return
	}

	command := message[0]
	payload := RemoveStringFromSlice(message, command)

	if command == "enable"{
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<enable> requires at least one argument")
			return
		}
		h.EnableCommand(payload, s, m)
		return
	}
	if command == "disable"{
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<disable> requires at least one argument")
			return
		}
		h.DisableCommand(payload, s, m)
		return

	}
	//TODO
	if command == "groups"{
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<group> requires at least one argument")
			return
		}
		h.ReadGroups(payload, s, m)
		return

	}
	if command == "users"{
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<user> requires at least one argument")
			return
		}
		h.ReadUsers(payload, s, m)
		return

	}
	if command == "channels"{
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<channels> requires at least one argument")
			return
		}
		h.ReadChannels(payload, s, m)
		return

	}

	if command == "usage"{
		if len(message) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<usage> requires at least one argument")
			return
		}
		h.DisplayUsage(payload, s, m)
		return
	}
	if command == "description"{
		if len(message) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<description> requires at least one argument")
			return
		}
		h.DisplayDescription(payload, s, m)
		return
	}
	if command == "list" {
		page := 0
		if len(message) < 2 {
			page = 0
			h.ListCommands(page, s, m)
		} else {
			page, err := strconv.Atoi(message[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, message[1] + " is not a valid page number")
				return
			}
			if page > 0 {
				page = page - 1
			}
			h.ListCommands(page, s, m)
		}
	}
}

func (h *CommandHandler) ReadGroups(message []string, s *discordgo.Session, m *discordgo.MessageCreate){
	command := message[0]
	payload := RemoveStringFromSlice(message, command)

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, command + " requires an argument")
		return
	}
	// list
	if command == "list" {
		groups, err := h.registry.GetGroups(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if len(groups) < 1 {
			s.ChannelMessageSend(m.ChannelID, ":\nNo Group assignments found for " + payload[0])
			return
		}

		var formatted string
		for i, group := range groups {

			if i == len(groups)-1 {
				formatted = formatted + group
			} else {
				formatted = formatted + group + ", "
			}
		}

		s.ChannelMessageSend(m.ChannelID, ":\nGroups for " + payload[0] + " : " + formatted )
		return
	}
	// add
	if command == "add" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, command + " requires two arguments <group> <command>")
			return
		}

		err := h.registry.AddGroup(payload[1], payload[0])
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, payload[1] + " has been added to the " + payload[0] + " group.")
		return
	}
	// remove
	if command == "remove" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, command + " requires two arguments <group> <command>")
			return
		}
		err := h.registry.RemoveGroup(payload[1], payload[0])
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, payload[1] + " has been removed from the " + payload[0] + " group.")
		return
	}
}

func (h *CommandHandler) ReadUsers(message []string, s *discordgo.Session, m *discordgo.MessageCreate){
	command := message[0]
	payload := RemoveStringFromSlice(message, command)


	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, command + " requires an argument")
		return
	}
	// list
	if command == "list" {
		users, err := h.registry.GetUsers(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if len(users) < 1 {
			s.ChannelMessageSend(m.ChannelID, ":\nNo User assignments found for " + payload[0])
			return
		}

		var formatted string
		for i, user := range users {
			dguser, err := s.User(user)
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			if i == len(users)-1 {
				formatted = formatted + dguser.Mention()
			} else {
				formatted = formatted + dguser.Mention() + ", "
			}
		}

		s.ChannelMessageSend(m.ChannelID, ":\nUsers for " + payload[0] + " : " + formatted )
		return
	}
	// add
	if command == "add" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, command + " requires two arguments <user> <command>")
			return
		}
		mentions := m.Mentions
		if len(mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Invalid Usage: User must be mentioned")
			return
		}

		err := h.registry.AddUser(payload[1], mentions[0].ID)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, mentions[0].Mention() + " has been added to " + payload[1])
		return
	}
	// remove
	if command == "remove" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, command + " requires two arguments <user> <command>")
			return
		}
		mentions := m.Mentions
		if len(mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Invalid Usage: User must be mentioned")
			return
		}
		err := h.registry.RemoveUser(payload[1], mentions[0].ID)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, mentions[0].Mention() + " has been removed from " + payload[0])
		return
	}
}

func (h *CommandHandler) ReadChannels(message []string, s *discordgo.Session, m *discordgo.MessageCreate){
	command := message[0]
	payload := RemoveStringFromSlice(message, command)


	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, command + " requires an argument")
		return
	}
	// list
	if command == "list" {
		channels, err := h.registry.ChannelList(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if len(channels) < 1 {
			s.ChannelMessageSend(m.ChannelID, ":\nNo Channel assignments found for " + payload[0])
			return
		}

		var formatted string
		for i, channel := range channels {
			dgchannel, err := s.Channel(channel)
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			if i == len(channels)-1 {
				formatted = formatted + "<#"+dgchannel.ID+">"
			} else {
				formatted = formatted + "<#"+dgchannel.ID+">, "
			}
		}

		s.ChannelMessageSend(m.ChannelID, ":\nChannels for " + payload[0] + " : " + formatted )
		return
	}
	// add
	if command == "add" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, command + " requires two arguments <channel> <command>")
			return
		}

		channelid := CleanChannel(payload[0])
		_, err := strconv.Atoi(channelid)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}

		channelmention, err := s.Channel(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		if channelmention.Name == "" {
			s.ChannelMessageSend(m.ChannelID, "Error during channel validation")
			return
		}

		err = h.registry.AddChannel(payload[1], channelid)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "<#"+channelmention.ID+">" + " has been added to " + payload[1])
		return

	}
	// remove
	if command == "remove" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, command + " requires two arguments <channel> <command>")
			return
		}

		channelid := CleanChannel(payload[0])
		_, err := strconv.Atoi(channelid)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}


		channelmention, err := s.Channel(channelid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		if channelmention.Name == "" {
			s.ChannelMessageSend(m.ChannelID, "Invalid Usage: Channel must be mentioned")
			return
		}

		err = h.registry.RemoveChannel(payload[1], channelid)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "<#"+channelmention.ID+">" + " has been removed from " + payload[0])
		return
	}
}

func (h *CommandHandler) EnableCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	h.registry.AddChannel(message[0], m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, "Command " + message[0] + " enabled for this channel")
	return
}

func (h *CommandHandler) DisableCommand(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	h.registry.RemoveChannel(message[0], m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, "Command " + message[0] + " disabled for this channel")
	return
}

func (h *CommandHandler) DisplayUsage(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, err := h.registry.GetCommand(message[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	formattedmessage := ":\n" + " Usage guide for " + message[0] + "\n"
	formattedmessage = formattedmessage + "```"
	formattedmessage = formattedmessage + command.Usage
	formattedmessage = formattedmessage + "```"

	s.ChannelMessageSend(m.ChannelID, formattedmessage)
	return
}

func (h *CommandHandler) DisplayDescription(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, err := h.registry.GetCommand(message[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	formattedmessage := ":\n" + " Description for " + message[0] + "\n"
	formattedmessage = formattedmessage + "```"
	formattedmessage = formattedmessage + command.Description
	formattedmessage = formattedmessage + "```"

	s.ChannelMessageSend(m.ChannelID, formattedmessage)
	return
}


func (h *CommandHandler) ListCommands(page int, s *discordgo.Session, m *discordgo.MessageCreate) {

	recordlist, err := h.registry.CommandsForChannel(page, m.ChannelID)
	if err != nil{

		if err.Error() == "not found" {
			s.ChannelMessageSend(m.ChannelID, "No commands for this channel found")
			return
		}
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	pagecount, err := h.registry.CommandsForChannelPageCount(m.ChannelID)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	formattedmessage := h.FormatCommands(recordlist) + "\nPage " + strconv.Itoa(page+1) +
		" of " + strconv.Itoa(pagecount)
	s.ChannelMessageSend(m.ChannelID, formattedmessage)

	return
}



func (h *CommandHandler) FormatCommands(recordlist []CommandRecord) (formattedlist string) {


	formattedmessage := ":\n" + " Command list for this channel (see command <description> or command <usage> if table is too small)\n"

	formattedmessage = formattedmessage + "```"
	formattedmessage = formattedmessage + "|---------------------|-------------------------|-------------------------|\n"
	formattedmessage = formattedmessage + "|       Command       |       Description       |          Usage          |\n"
	formattedmessage = formattedmessage + "|---------------------|-------------------------|-------------------------|\n"


	// I am aware this is probably overcomplicated, but it's not a very frequently used command
	commandColumnlen := len("|       Command       ")-2
	descriptionColumnlen := len("|       Description       ")-2
	usageColumnlen := len("|          Usage          ")-2

	for _, command := range recordlist {
		commandlen := len(command.Command)
		Command := command.Command

		if commandlen < commandColumnlen {
			diff := commandColumnlen - commandlen
			for i := 0; i < diff ; i++ {
				Command = Command + " "
			}
		}

		if commandlen > commandColumnlen {
			diff := commandlen - commandColumnlen + 1
			for i := 0; i < diff ; i++ {
				Command = Command[:len(Command)-1]
			}
			Command = Command + " "
		}

		descriptionlen := len(command.Description)
		Description := command.Description

		if descriptionlen < descriptionColumnlen {
			diff := descriptionColumnlen - descriptionlen
			for i := 0; i < diff ; i++ {
				Description = Description + " "
			}
		}

		if descriptionlen > descriptionColumnlen {
			diff := descriptionlen - descriptionColumnlen + 1
			for i := 0; i < diff ; i++ {
				Description = Description[:len(Description)-1]
			}
			Description = Description + " "
		}

		usagelen := len(command.Usage)
		Usage := command.Usage

		if usagelen < usageColumnlen {
			diff := usageColumnlen - usagelen
			for i := 0; i < diff ; i++ {
				Usage = Usage + " "
			}
		}

		if usagelen > usageColumnlen {
			diff := usagelen - usageColumnlen + 1
			for i := 0; i < diff ; i++ {
				Usage = Usage[:len(Usage)-1]
			}
			Usage = Usage + " "
		}

		formattedmessage = formattedmessage + "| " + Command + "| " + Description + "| " + Usage +"|\n"
	}

	formattedmessage = formattedmessage + "|---------------------|-------------------------|-------------------------|\n"
	formattedmessage = formattedmessage + "```"

	return formattedmessage

}


