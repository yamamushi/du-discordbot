package main

// Utility Functions

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	//"strconv"
	"strconv"
)

// minDuration and maxDuration const for rounding
const (
	minTimeDuration time.Duration = -1 << 63
	maxTimeDuration time.Duration = 1<<63 - 1
)

// RemoveStringFromSlice function
func RemoveStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// RemoveStringFromSlice function
func RemoveIntFromSlice(s []int, r int) []int {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// SafeInput function
func SafeInput(s *discordgo.Session, m *discordgo.MessageCreate, conf *Config) bool {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return false
	}

	// Ignore bots
	if m.Author.Bot {
		return false
	}

	if !strings.HasPrefix(m.Content, conf.DUBotConfig.CP) {
		return false
	}

	// Set our command prefix to the default one within our config file
	message := strings.Fields(m.Content)
	if len(message) < 1 {
		return false
	}

	return true
}

// CleanCommand function
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

// SplitPayload function
func SplitPayload(input []string) (command string, message []string) {

	// Remove the prefix from our command
	command = input[0]
	message = RemoveStringFromSlice(input, command)

	return command, message

}

// SplitCommandFromArg function
func SplitCommandFromArgs(input []string) (command string, message string) {

	// Remove the prefix from our command
	command = input[0]
	payload := RemoveStringFromSlice(input, command)

	for _, value := range payload {
		message = message + value + " "
	}
	return command, message
}

// RemoveFromString function
func RemoveFromString(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

// CleanChannel function
func CleanChannel(mention string) string {

	mention = strings.TrimPrefix(mention, "<#")
	mention = strings.TrimSuffix(mention, ">")
	mentionSplit := strings.Split(mention, "-")
	return mentionSplit[0]

}

// MentionChannel function
func MentionChannel(channelid string, s *discordgo.Session) (mention string, err error) {
	dgchannel, err := s.Channel(channelid)
	if err != nil {
		return "", err
	}

	return "<#" + dgchannel.ID + ">", nil
}

// CheckPermissions function
func CheckPermissions(command string, channelid string, user *User, s *discordgo.Session, com *CommandHandler) bool {

	usergroups, err := com.user.GetGroups(user.ID)
	if err != nil {
		//fmt.Println("Error Retrieving User Groups for " + user.ID)
		return false
	}

	commandgroups, err := com.registry.GetGroups(command)
	if err != nil {
		//fmt.Println("Error Retrieving Registry Groups for " + command)
		return false
	}

	commandchannels, err := com.registry.GetChannels(command)
	if err != nil {
		//fmt.Println("Error Retrieving Channels for " + command)
		return false
	}

	commandusers, err := com.registry.GetUsers(command)
	if err != nil {
		//fmt.Println("Error Retrieving Users for " + command)
		return false
	}

	// Verify our channel is valid
	_, err = s.Channel(channelid)
	if err != nil {
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

// MentionOwner function
func MentionOwner(conf *Config, s *discordgo.Session, m *discordgo.MessageCreate) (mention string, err error) {
	user, err := s.User(conf.DiscordConfig.AdminID)
	if err != nil {
		return mention, err
	}

	return user.Mention(), nil
}

// OwnerName function
func OwnerName(conf *Config, s *discordgo.Session, m *discordgo.MessageCreate) (name string, err error) {
	user, err := s.User(conf.DiscordConfig.AdminID)
	if err != nil {
		return name, err
	}

	return user.Username, nil
}

// IsVoiceChannelEmpty function
func IsVoiceChannelEmpty(s *discordgo.Session, channelid string, botid string) bool {

	channel, err := s.Channel(channelid)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if len(guild.VoiceStates) > 0 {
		for _, state := range guild.VoiceStates {
			if state.ChannelID == channelid && state.UserID != botid {
				return false
			}
		}
		return true
	}

	return true

}

// TruncateTime returns the result of rounding d toward zero to a multiple of m.
// If m <= 0, Truncate returns d unchanged.
func TruncateTime(d time.Duration, m time.Duration) time.Duration {
	if m <= 0 {
		return d
	}
	return d - d%m
}

// RoundTime returns the result of rounding d to the nearest multiple of m.
// The rounding behavior for halfway values is to round away from zero.
// If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration,
// Round returns the maximum (or minimum) duration.
// If m <= 0, Round returns d unchanged.
func RoundTime(d time.Duration, m time.Duration) time.Duration {
	if m <= 0 {
		return d
	}
	r := d % m
	if d < 0 {
		r = -r
		if r+r < m {
			return d + r
		}
		if d1 := d - m + r; d1 < d {
			return d1
		}
		return minTimeDuration // overflow
	}
	if r+r < m {
		return d - r
	}
	if d1 := d + m - r; d1 > d {
		return d1
	}
	return maxTimeDuration // overflow
}

// truncateString function Shorten a string to num characters
func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}

//
func MessageHasMeme(message string, meme string) bool {
	if strings.HasPrefix(message, meme) {
		return true
	}
	if strings.HasSuffix(message, meme) {
		return true
	}
	if strings.Contains(message, " "+meme+" ") {
		return true
	}
	if strings.Contains(message, " "+meme+"-") || strings.Contains(message, " "+meme+",") || strings.Contains(message, " "+meme+".") {
		return true
	}
	if strings.Contains(message, " "+meme+":") || strings.Contains(message, " "+meme+";") || strings.Contains(message, " "+meme+"+") {
		return true
	}
	if strings.Contains(message, " "+meme+"=") || strings.Contains(message, " "+meme+"?") || strings.Contains(message, " "+meme+"/") {
		return true
	}
	if strings.Contains(message, " "+meme+"!") || strings.Contains(message, " "+meme+"'") || strings.Contains(message, " "+meme+"\"") {
		return true
	}
	return false
}

func VerifyNDAChannel(channelID string, conf *Config) bool {
	if channelID == conf.RolesConfig.NDAChannelID {
		return true
	}

	return false
}

func CreateDirIfNotExist(dir string) (err error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// getRoleIDByName function
func getRoleIDByName(s *discordgo.Session, guildID string, name string) (roleid string, err error) {
	name = strings.Title(name)
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if role.Name == name {
			return role.ID, nil
		}
	}
	return "", errors.New("Role ID Not Found: " + name)
}

// getChannelIDByName function
func getChannelIDByName(s *discordgo.Session, guildID string, name string) (roleid string, err error) {
	//name = strings.Title(name)
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.Name == name {
			return channel.ID, nil
		}
	}
	return "", errors.New("Channel ID Not Found: " + name)
}

func GetMemberList(s *discordgo.Session, conf *Config) ([]*discordgo.Member, error) {

	guild, err := s.Guild(conf.DiscordConfig.GuildID)
	if err != nil {
		return nil, err
	}
	return guild.Members, nil

}

func SendFileToChannel(path string, message string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	if _, err := os.Stat(path); err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	pathComplex := strings.Split(path, "/")
	fileName := pathComplex[len(pathComplex)-1]

	ms := &discordgo.MessageSend{
		Content: message,
		Files: []*discordgo.File{
			&discordgo.File{
				Name:   fileName,
				Reader: f,
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, ms)
	if err != nil {
		return err
	}
	return nil
}

func ParseDuration(duration string) (convertedTime time.Duration, totalminutes int64, err error) {

	daysstring := "0"
	hoursstring := "0"
	minutesstring := "0"
	var days, hours, minutes int64

	separated := strings.Split(duration, " ")

	for _, field := range separated {

		for _, value := range field {
			switch {
			case value >= '0' && value <= '9':
				if strings.Contains(field, "d") {
					daysstring = strings.TrimSuffix(field, "d")
					days, err = strconv.ParseInt(daysstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse days")
					}
				} else if strings.Contains(field, "h") {
					hoursstring = strings.TrimSuffix(field, "h")
					hours, err = strconv.ParseInt(hoursstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse hours")
					}
				} else if strings.Contains(field, "m") {
					minutesstring = strings.TrimSuffix(field, "m")
					minutes, err = strconv.ParseInt(minutesstring, 10, 64)
					if err != nil {
						return 0, 0, errors.New("Could not parse minutes")
					}
				} else {
					return 0, 0, errors.New("Invalid time interval format")
				}
				break
			default:
				return 0, 0, errors.New("Invalid time interval format")
			}
			break
		}
	}

	if days == 0 && hours == 0 && minutes == 0 {
		return 0, 0, errors.New("Invalid interval specified")
	}

	totalminutes = (days * 24 * 60) + (hours * 60) + minutes
	convertedTime = time.Duration(totalminutes * 60 * 1000 * 1000 * 1000)
	return convertedTime, totalminutes, nil
}

// getRoleNameByID function
func getRoleNameByID(roleID string, guildID string, s *discordgo.Session) (rolename string, err error) {

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return "", err
	}

	for _, role := range roles {

		if role.ID == roleID {
			return role.Name, nil
		}
	}

	return "", errors.New("Role " + roleID + " not found in guild " + guildID)
}

// FlushMessages function
func FlushMessages(s *discordgo.Session, channelID string, count int) (err error) {

	if count <= 0 {
		return errors.New(":rotating_light: Invalid message count supplied")
	}

	if count > 100 {
		return errors.New(":rotating_light: Count must be less than or equal to 100")

	}

	_, err = s.Channel(channelID)
	if err != nil {
		return err
	}

	messages, err := s.ChannelMessages(channelID, count, "", "", "")
	if err != nil {
		return err
	}

	if count > len(messages) {
		count = len(messages)
	}

	var flushmessages []string

	for i := 0; i < count; i++ {
		flushmessages = append(flushmessages, messages[i].ID)
	}

	err = s.ChannelMessagesBulkDelete(channelID, flushmessages)
	if err != nil {
		return err
	}

	return nil
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	if m > 0 {
		return fmt.Sprintf("%02d hours %02d minutes", h, m)
	} else {
		return fmt.Sprintf("%02d hours", h)
	}
}

func AppendStringIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

// Contains tells whether a contains x.
func StringSliceContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}

func IsReachableURL(toTest string) bool {
	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	_, err := client.Get(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}
