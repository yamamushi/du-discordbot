package main

// Utility Functions

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
)


func RemoveStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}


func SafeInput(s *discordgo.Session, m *discordgo.MessageCreate, conf *Config) bool {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return false
	}

	// Ignore bots
	if m.Author.Bot {
		return false
	}

	if !strings.HasPrefix(m.Content, conf.DUBotConfig.CP){
		return false
	}

	// Set our command prefix to the default one within our config file
	message := strings.Fields(m.Content)
	if len(message) < 1 {
		return false
	}

	return true
}


func CleanCommand(input string, conf *Config) (command string, message []string) {

	// Set our command prefix to the default one within our config file
	cp := conf.DUBotConfig.CP
	message = strings.Fields(input)

	// Remove the prefix from our command
	message[0] = strings.Trim(message[0], cp)
	command = message[0]
	message = RemoveStringFromSlice(message, command)

	return command, message

}


func SplitPayload(input []string) (command string, message []string) {

	// Remove the prefix from our command
	command = input[0]
	message = RemoveStringFromSlice(input, command)

	return command, message

}

func RemoveFromMessage(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func CleanChannel(mention string) (string){

	mention = strings.TrimPrefix(mention, "<#")
	mention = strings.TrimSuffix(mention, ">")
	return mention

}


func MentionChannel(channelid string, s *discordgo.Session) (mention string, err error){
	dgchannel, err := s.Channel(channelid)
	if err != nil{
		return "", err
	}

	return "<#"+dgchannel.ID+">", nil
}


func CheckPermissions(command string, channelid string, user *User, s *discordgo.Session, com *CommandHandler) bool {

	usergroups, err := com.user.GetGroups(user.ID)
	if err != nil{
		//fmt.Println("Error Retrieving User Groups for " + user.ID)
		return false
	}

	commandgroups, err := com.registry.GetGroups(command)
	if err != nil{
		//fmt.Println("Error Retrieving Registry Groups for " + command)
		return false
	}

	commandchannels, err := com.registry.GetChannels(command)
	if err != nil{
		//fmt.Println("Error Retrieving Channels for " + command)
		return false
	}

	commandusers, err := com.registry.GetUsers(command)
	if err != nil{
		//fmt.Println("Error Retrieving Users for " + command)
		return false
	}

	// Verify our channel is valid
	_, err = s.Channel(channelid)
	if err != nil{
		return false
	}

	// Look to see if the provided channel id matches one in the command's channel list
	match := false
	for _, commandchannelid := range commandchannels {
		if commandchannelid == channelid {
			match = true
		}
	}
	// If command is not in channel list we return false
	if !match {
		return false
	}

	// Look to see if our user ID is in the users list for the command.
	for _, commanduser := range commandusers {
		if commanduser == user.ID {
			return true
		}
	}

	// Finally we want to try to check the user group list
	for _, usergroup := range usergroups {
		for _, commandgroup := range commandgroups {
			if usergroup == commandgroup {
				return true
			}
		}
	}

	return false
}


func MentionOwner(conf *Config, s *discordgo.Session, m *discordgo.MessageCreate) (mention string, err error) {
	user, err := s.User(conf.DiscordConfig.AdminID)
	if err != nil {
		return mention, err
	}

	return user.Mention(), nil
}

func OwnerName(conf *Config, s *discordgo.Session, m *discordgo.MessageCreate) (name string, err error) {
	user, err := s.User(conf.DiscordConfig.AdminID)
	if err != nil {
		return name, err
	}

	return user.Username, nil
}


func IsVoiceChannelEmpty(s *discordgo.Session, channelid string, botid string)(bool){

	channel, err := s.Channel(channelid)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	guild, err := s.Guild(channel.GuildID)
	if err != nil{
		fmt.Println(err.Error())
		return false
	}

	if len(guild.VoiceStates) > 0 {
		for _, state := range guild.VoiceStates {
			if state.ChannelID == channelid  && state.UserID != botid{
				return false
			}
		}
		return true
	} else {
		return true
	}
}