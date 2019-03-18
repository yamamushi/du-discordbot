package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/mgo.v2"
	"reflect"
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
	callback    *CallbackHandler

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

	if !user.Admin {
		return errors.New("You do not have permission to use this command!")
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
		return errors.New("Info page for \"" + recordname + "\" not found")
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

	if !user.Admin {
		return errors.New("You do not have permission to use this command!")
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
		return errors.New("Record \""+ recordname + "\" already exists")
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

func (h *InfoHandler) EditMenu(recordname string, messageID string, channelID string, userID string, s *discordgo.Session) (err error){

	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		return err
	}
	defer session.Close()

	record, err := h.infodb.GetRecordFromDB(recordname, *collection)
	if err != nil {
		return err
	}

	err = s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor"
	embed.Description = "Currently Editing \""+record.Name+"\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}

	fields := []*discordgo.MessageEmbedField{}

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Set the record type"
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Set the record description"
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "Set the record image"
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	optionfour := discordgo.MessageEmbedField{}
	optionfour.Name = "4⃣"
	optionfour.Value = "Set the record color"
	optionfour.Inline = true
	fields = append(fields, &optionfour)

	optionfive := discordgo.MessageEmbedField{}
	optionfive.Name = "5⃣"
	optionfive.Value = "Configure record properties"
	optionfive.Inline = true
	fields = append(fields, &optionfive)

	optionsix := discordgo.MessageEmbedField{}
	optionsix.Name = "6⃣"
	optionsix.Value = "Close this editing session"
	optionsix.Inline = true
	fields = append(fields, &optionsix)

	embed.Fields = fields

	/*
		if record.RecordType == "element" {
			return h.RenderElementPage(record, s, m)
		} else if record.RecordType == "location" {
			return h.RenderLocationPage(record, s, m)
		} else if record.RecordType == "resource" {
			return h.RenderResourcePage(record, s, m)
		} else if record.RecordType == "skill" {
			return h.RenderSkillPage(record, s, m)
		}
	 */

    var reactions []string
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	reactions = append(reactions, "3⃣")
	reactions = append(reactions, "4⃣")
	reactions = append(reactions, "5⃣")
	reactions = append(reactions, "6⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleEditorMainMenu, messageID, channelID, userID, recordname, s)
	return nil
}



func (h *InfoHandler) HandleEditorMainMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	defer session.Close()

	record, err := h.infodb.GetRecordFromDB(recordname, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}



	if reaction == "1⃣" {
		h.SetRecordTypeMenu(record, s, m)
		return
	}
	if reaction == "2⃣" {
	}
	if reaction == "3⃣" {
	}
	if reaction == "4⃣" {
	}
	if reaction == "5⃣" {
	}
	if reaction == "6⃣" {
	}

	return
}

func (h *InfoHandler) SetRecordTypeMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Record Type Selection"
	embed.Description = "Currently Editing \""+record.Name+"\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}

	fields := []*discordgo.MessageEmbedField{}

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Satellite"
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Element"
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "Resource"
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	optionfour := discordgo.MessageEmbedField{}
	optionfour.Name = "4⃣"
	optionfour.Value = "Skill"
	optionfour.Inline = true
	fields = append(fields, &optionfour)


	var reactions []string
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	reactions = append(reactions, "3⃣")
	reactions = append(reactions, "4⃣")
	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}


	_, err := s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleSetRecordType, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetRecordType(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	defer session.Close()

	record, err := h.infodb.GetRecordFromDB(recordname, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	if reaction == "1⃣" {
		record.RecordType = "satellite"
	}
	if reaction == "2⃣" {
		record.RecordType = "element"
	}
	if reaction == "3⃣" {
		record.RecordType = "resource"
	}
	if reaction == "4⃣" {
		record.RecordType = "skill"
	}

	err = h.infodb.SaveRecordToDB(record, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error Saving Record: " + err.Error())
		return
	}

	err = h.EditMenu(record.Name, messageID, channelID, userID, s)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	return
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

	if record.RecordType == "element" {
		return h.RenderElementPage(record, s, m)
	} else if record.RecordType == "location" {
		return h.RenderLocationPage(record, s, m)
	} else if record.RecordType == "resource" {
		return h.RenderResourcePage(record, s, m)
	} else if record.RecordType == "skill" {
		return h.RenderSkillPage(record, s, m)
	}

	return errors.New("Record type not known")
}

func (h *InfoHandler) RenderElementPage(record InfoRecord, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	embed := &discordgo.MessageEmbed{}
	embed.Title = record.Name

	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return nil
}

func (h *InfoHandler) RenderLocationPage(record InfoRecord, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
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
