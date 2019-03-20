package main

import (
	"github.com/bwmarrin/discordgo"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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
	embed.Description = "Currently Editing \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField

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
		h.SetRecordTypeMenu(record, s, m)
		return
	}
	if reaction == "2⃣" {
		h.SetRecordDescriptionMenu(record, s, m)
		return
	}
	if reaction == "3⃣" {
		h.SetRecordImageURLMenu(record, s, m)
		return
	}
	if reaction == "4⃣" {
		h.SetRecordColorMenu(record, s, m)
		return
	}
	if reaction == "5⃣" {
		if record.RecordType == "" {
			errormessage, _ := s.ChannelMessageSend(channelID, "Error: Cannot set record properties without record type, please configure this first.")
			time.Sleep(10*time.Second)
			_ = s.ChannelMessageDelete(channelID, errormessage.ID)
			h.SetRecordTypeMenu(record, s, m)
			return
		}
		if record.RecordType == "satellite" {
			h.SatellitePropertiesMenu(record, s, m)
			return
		} else {
			return
		}

	}
	if reaction == "6⃣" {
		err = s.ChannelMessageDelete(channelID, messageID)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		s.ChannelMessageSend(channelID, "Info Editor Session Closed")
		return
	} else {
		h.reactions.Watch(h.HandleEditorMainMenu, messageID, channelID, userID, recordname, s)
		return
	}
}

// Record Type
func (h *InfoHandler) SetRecordTypeMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Record Type Selection"
	embed.Description = "Currently Editing \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

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
	reactions = append(reactions, "⬅")
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

	if reaction == "⬅" {
		err = h.EditMenu(record.Name, messageID, channelID, userID, s)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
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


// Record Description

// SetRecordDescriptionMenu
func (h *InfoHandler) SetRecordDescriptionMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Record Description"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	if record.Description != "" {
		optionone := discordgo.MessageEmbedField{}
		optionone.Name = "Current Description"
		optionone.Value = record.Description
		optionone.Inline = false
		fields = append(fields, &optionone)
	}

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new description below or select ⬅ to return to the main menu"
	optiontwo.Inline = false
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "⬅")

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}


	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.infocallback.Watch(h.HandleSetRecordDescription, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetRecordDescriptionReactions, messageID, channelID, userID, record.Name, s)
	return
}

// HandleSetRecordDescription function
func (h *InfoHandler) HandleSetRecordDescription(recordname string, userID string, description string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	//h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid descriptions
	if len(description) > 2047 {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided description was too long, " +
			"must be shorter than 2048 characters for discord embed compatibility. Returning to previous menu.")
		h.SetRecordDescriptionMenu(record, s, m)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Record Description"
	embed.Description = "Confirm Description Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Description"
	optionone.Value = description
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the description menu"
	optiontwo.Inline = false
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "✅")
	reactions = append(reactions, "❎")

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	payload := recordname + "|#|" + description
	h.reactions.Watch(h.HandleSetRecordDescriptionConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetRecordDescriptionReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {


	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
		err := h.EditMenu(recordname, messageID, channelID, userID, s)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		return
	}
	h.reactions.Watch(h.HandleSetRecordDescriptionReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetRecordDescriptionConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetRecordDescriptionConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	description := payload[1]

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

	if reaction == "✅" {

		record.Description = description
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetRecordDescriptionMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to description menu
		h.SetRecordDescriptionMenu(record, s, m)
		return
	}
}


// Record Image URL

func (h *InfoHandler) SetRecordImageURLMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Record Image URL"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new url below or select ⬅ to return to the main menu"
	optiontwo.Inline = false
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "⬅")

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}


	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.infocallback.Watch(h.HandleSetRecordImageURL, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetRecordImageURLReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetRecordImageURL(recordname string, userID string, imageurl string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	//h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid urls
	_, err = url.ParseRequestURI(imageurl)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided url is invalid, please check your entry and try again:\n```" + imageurl+"```")
		r := discordgo.MessageReaction{ChannelID:channelID, MessageID: messageID, UserID: userID}
		h.SetRecordImageURLMenu(record, s, r)
		time.Sleep(15*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Record Image URL"
	embed.Description = "Confirm Image URL Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:imageurl}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Image URL"
	optionone.Value = imageurl
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Image URL menu"
	optiontwo.Inline = false
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "✅")
	reactions = append(reactions, "❎")

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	payload := recordname + "|#|" + imageurl
	h.reactions.Watch(h.HandleSetRecordImageURLConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetRecordImageURLReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
		err := h.EditMenu(recordname, messageID, channelID, userID, s)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetRecordImageURLReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetRecordImageURLConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetRecordImageURLConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	imageurl := payload[1]

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

	if reaction == "✅" {

		record.ImageURL = imageurl
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetRecordImageURLMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu
		h.SetRecordImageURLMenu(record, s, m)
		return
	}
}


// Record Color

func (h *InfoHandler) SetRecordColorMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Record Color"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new color code below or select ⬅ to return to the main menu"
	optiontwo.Inline = false
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "⬅")

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}


	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.infocallback.Watch(h.HandleSetRecordColor, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetRecordColorReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetRecordColor(recordname string, userID string, color string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	//h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid colors
	colorcode, err := strconv.Atoi(color)
	if err != nil || colorcode > 16777215 {
		errormessage, _ := s.ChannelMessageSend(channelID,
			"Provided color code is invalid, please check your entry and try again:\n```" + color+"```")
		r := discordgo.MessageReaction{ChannelID:channelID, MessageID: messageID, UserID: userID}
		h.SetRecordColorMenu(record, s, r)
		time.Sleep(15*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Record Color"
	embed.Description = "Confirm Color Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = colorcode

	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Image Color"
	optionone.Value = color
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Color menu"
	optiontwo.Inline = false
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "✅")
	reactions = append(reactions, "❎")

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	payload := recordname + "|#|" + color
	h.reactions.Watch(h.HandleSetRecordColorConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetRecordColorReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
		err := h.EditMenu(recordname, messageID, channelID, userID, s)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetRecordColorReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetRecordColorConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetRecordColorConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	color := payload[1]

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

	if reaction == "✅" {

		// We check again for consistency
		colorcode, err := strconv.Atoi(color)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			h.SetRecordImageURLMenu(record, s, m)
			return
		}

		record.Color = colorcode
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetRecordColorMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu
		h.SetRecordColorMenu(record, s, m)
		return
	}
}


// Record Properties (yuck)
// This is going to be a nightmare of code...
