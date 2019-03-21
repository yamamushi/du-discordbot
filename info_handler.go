package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/mgo.v2"
	"strings"
	"time"
)

type InfoHandler struct {

	conf        *Config
	registry    *CommandRegistry
	db          *DBHandler
	userdb      *UserHandler
	infodb      *InfoDBInterface
	reactions   *InfoReactionsHandler
	infocallback    *InfoCallbackHandler

}


// Init function
func (h *InfoHandler) Init() {
	_ = h.RegisterCommands()
	h.infodb = &InfoDBInterface{db:h.db, conf:h.conf}
}

// RegisterCommands function
func (h *InfoHandler) RegisterCommands() (err error) {
	h.registry.Register("info", "Alpha Member Info Utility", "See help page ```~info help```")
	return nil
}

// Read function
func (h *InfoHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"info") {
		if h.registry.CheckPermission("info", m.ChannelID, user) {

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
	// Else
}


// ParseCommand function
func (h *InfoHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)
	for i, _ := range payload {
		payload[i] = strings.ToLower(payload[i])
	}
	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command "+command+" expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "edit" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "edit expects a record name")
			return
		}
		err := h.Edit(payload, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		return
	}
	if payload[0] == "new" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "new expects a record name")
			return
		}
		err := h.New(payload, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
			return
		}
		return
	}

	recordname := strings.Join(payload, " ")

	err := h.RenderInfoPage(recordname, s, m)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
		return
	}
}

func (h *InfoHandler) GetMongoCollecton() (collection *mgo.Collection, session *mgo.Session, err error) {

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return &mgo.Collection{}, &mgo.Session{}, err
	}
	//defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	collection = session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.InfoRecordColumn)

	return collection, session, nil
}

func (h *InfoHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate) {

	output := ":bulb: Info System Usage: \n```\n" +
		"~info help - display this help screen\n" +
		"~location <location> - sets your location (this will result in location system notifications, be aware of what this means)\n" +
		"~location reset - resets your location status (this will disable notifications)\n" +
		"~info <resource> - displays information about a named resource\n" +
		"~info <location> - displays information about a named location\n" +
		"~info <element> - displays information about a given element\n" +
		"~info edit <name> - Opens the editor menu for a given record\n" +
		"~info new <name> - Creates a new record and opens the editor menu for it\n" +
		"```\n"

	s.ChannelMessageSend(m.ChannelID, output)
	return

}

func (h  *InfoHandler) Edit(args []string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	user, err := h.userdb.GetUser(m.Author.ID)
	if err != nil {
		return err
	}

	if !user.Editor {
		return errors.New("you do not have permission to use this command")
	}

	// Remove first element from argument list (which is "edit" here)
	args = append(args[:0], args[1:]...)

	recordname := strings.Join(args, " ")
	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = h.infodb.GetRecordFromDB(recordname, *collection)
	if err != nil {
		return errors.New("Info page for \"**" + strings.Title(recordname) + "**\" not found")
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor"

	rootmessage, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	err = h.EditMenu(recordname, rootmessage.ID, m.ChannelID, m.Author.ID, s)
	if err != nil {
		return err
	}

	return nil
}

func (h  *InfoHandler) New(args []string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	user, err := h.userdb.GetUser(m.Author.ID)
	if err != nil {
		return err
	}

	if !user.Editor {
		return errors.New("you do not have permission to use this command")
	}

	// Setup our db session
	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		return err
	}
	defer session.Close()

	// Remove first element from argument list (which is "edit" here)
	args = append(args[:0], args[1:]...)


	recordname := strings.Join(args, " ")
	_, err = h.infodb.GetRecordFromDB(recordname, *collection)
	if err == nil {
		return errors.New("Record \"**"+ strings.ToTitle(recordname) + "**\" already exists")
	}

	err = h.infodb.NewInfoRecord(recordname, *collection)
	if err != nil {
		return err
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor"

	rootmessage, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	err = h.EditMenu(recordname, rootmessage.ID, m.ChannelID, m.Author.ID, s)
	if err != nil {
		return err
	}

	return nil
}


func (h *InfoHandler) RenderInfoPage(recordname string, s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		return err
	}
	defer session.Close()

	record, err := h.infodb.GetRecordFromDB(recordname, *collection)
	if err != nil {
		return errors.New("Info page for \"" + recordname + "\" not found")
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = record.Name
	rootmessage, _ := s.ChannelMessageSendEmbed(m.ChannelID, embed)

	if record.RecordType == "element" {
		return h.RenderElementPage(record, s, m)
	} else if record.RecordType == "satellite" {
		r := &discordgo.MessageReaction{MessageID:rootmessage.ID, ChannelID:rootmessage.ChannelID, UserID:m.Author.ID}
		h.ViewSatelliteInfoMenu(record, s, r)
		return nil
	} else if record.RecordType == "resource" {
		return h.RenderResourcePage(record, s, m)
	} else if record.RecordType == "skill" {
		return h.RenderSkillPage(record, s, m)
	}

	return errors.New("record type not known")
}

func (h *InfoHandler) RenderElementPage(record InfoRecord, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	embed := &discordgo.MessageEmbed{}
	embed.Title = record.Name

	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return nil
}

func (h *InfoHandler) RenderResourcePage(record InfoRecord, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	embed := &discordgo.MessageEmbed{}
	embed.Title = record.Name

	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return nil
}

func (h *InfoHandler) RenderSkillPage(record InfoRecord, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	embed := &discordgo.MessageEmbed{}
	embed.Title = record.Name

	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return nil
}

func (h *InfoHandler) ValidatePosition(position string) (bool) {

	//::pos{0,2,0.7104,103.0054,-123.9859}

	if strings.HasPrefix(position, "::pos{") {
		position = strings.TrimPrefix(position, "::{")
	} else {
		return false
	}

	if strings.HasSuffix(position, "}") {
		position = strings.TrimSuffix(position, "}")
	} else {
		return false
	}

	coords := strings.Split(position, ",")
	if len(coords) != 5 {
		return false
	}

	return true
}

func (h *InfoHandler) SetUserLocation(userID string, location string, position string)(err error){

	if !h.ValidatePosition(position) && position != "confirm" && position != "space"{
		return errors.New("Invalid position")
	}

	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		return err
	}
	defer session.Close()

	record, err := h.infodb.GetRecordFromDB(location, *collection)
	if err != nil {
		return err
	}

	userrecord, err := h.infodb.GetRecordFromDB(userID, *collection)
	if err != nil {
		err = h.infodb.SaveRecordToDB(InfoRecord{Name:userID, RecordType:"user", User:UserRecord{UserID:userID}}, *collection)
		if err != nil {
			return err
		}
		userrecord, err = h.infodb.GetRecordFromDB(userID, *collection)
		if err != nil {
			return err
		}
	}

	userrecord.User.CurrentLocation = location
	if position != "confirm" && position != "space" {
		userrecord.Position = position
		record.Satellite.UserList = AppendStringIfMissing(record.Satellite.UserList, userID)
	}
	if position == "space" {
		userrecord.Position = "::pos{0,0,0,0,0}"
		userrecord.User.CurrentLocation = "space"
		record.Satellite.UserList = RemoveStringFromSlice(record.Satellite.UserList, userID)
	}
	if position == "confirm" {
		userrecord.Position = "::pos{0,0,0,0,0}"
		record.Satellite.UserList = AppendStringIfMissing(record.Satellite.UserList, userID)
	}
	

	err = h.infodb.SaveRecordToDB(record, *collection)
	if err != nil {
		return err
	}

	err = h.infodb.SaveRecordToDB(userrecord, *collection)
	if err != nil {
		return err
	}

	return nil

}
