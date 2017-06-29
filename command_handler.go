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
	// list

	// add

	// remove
}

func (h *CommandHandler) ReadUsers(message []string, s *discordgo.Session, m *discordgo.MessageCreate){
	// list

	// add

	// remove
}

func (h *CommandHandler) ReadChannels(message []string, s *discordgo.Session, m *discordgo.MessageCreate){
	// list

	// add

	// remove
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


