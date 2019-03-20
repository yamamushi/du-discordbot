package main

import (
	"github.com/bwmarrin/discordgo"
	"html"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Satellites
func (h *InfoHandler) SatellitePropertiesMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Properties :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Configure Satellite Type - Current Value: " + strings.Title(record.Satellite.SatelliteType)
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Configure Satellite Details"
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "Configure Satellite Elements"
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	optionfour := discordgo.MessageEmbedField{}
	optionfour.Name = "4⃣"
	optionfour.Value = "Configure Satellite Moons"
	optionfour.Inline = true
	fields = append(fields, &optionfour)

	optionfive := discordgo.MessageEmbedField{}
	optionfive.Name = "5⃣"
	optionfive.Value = "Configure Satellite Territories"
	optionfive.Inline = true
	fields = append(fields, &optionfive)
	/*
		optionsix := discordgo.MessageEmbedField{}
		optionsix.Name = "6⃣"
		optionsix.Value = "Configure Satellite Distances"
		optionsix.Inline = true
		fields = append(fields, &optionsix)
	*/
	embed.Fields = fields

	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	reactions = append(reactions, "3⃣")
	reactions = append(reactions, "4⃣")
	reactions = append(reactions, "5⃣")
	//	reactions = append(reactions, "6⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleSatellitePropertiesMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatellitePropertiesMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

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
		_ = h.EditMenu(recordname, messageID, channelID, userID, s)
		return
	}
	// Type
	if reaction == "1⃣" {
		h.SetSatelliteTypeMenu(record, s, m)
		return
	}
	// Details
	if reaction == "2⃣" {
		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// Elements
	if reaction == "3⃣" {
		h.SetSatelliteElementsMenu(record, s, m)
		return
	}
	// Moons
	if reaction == "4⃣" {
		return
	}
	// Territories
	if reaction == "5⃣" {
		h.SetSatelliteTerritoriesMenu(record, s, m)
		return
	} else {
		h.reactions.Watch(h.HandleSatellitePropertiesMenu, messageID, channelID, userID, recordname, s)
		return
	}
}


// Satellite Type

func (h *InfoHandler) SetSatelliteTypeMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Satellite Type Selection :satellite:"
	embed.Description = "Currently Editing \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Planet"
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Moon"
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)


	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
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
	h.reactions.Watch(h.HandleSetSatelliteType, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteType(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

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
		h.SatellitePropertiesMenu(record, s, m)
		return
	}
	if reaction == "1⃣" {
		record.Satellite.SatelliteType = "planet"
	}
	if reaction == "2⃣" {
		record.Satellite.SatelliteType = "moon"
	}

	err = h.infodb.SaveRecordToDB(record, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error Saving Record: " + err.Error())
		return
	}

	h.SatellitePropertiesMenu(record, s, m)
	return
}


// Details Menu

func (h *InfoHandler) SetSatelliteDetailsMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Details :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Discovered By: " + strings.Title(record.Satellite.DiscoveredBy)
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "System Zone: " + strings.Title(record.Satellite.SystemZone)
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "Atmosphere: " + strings.Title(record.Satellite.Atmosphere)
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	optionfour := discordgo.MessageEmbedField{}
	optionfour.Name = "4⃣"
	optionfour.Value = "Gravity: " + strings.Title(record.Satellite.Gravity)
	optionfour.Inline = true
	fields = append(fields, &optionfour)

	optionfive := discordgo.MessageEmbedField{}
	optionfive.Name = "5⃣"
	optionfive.Value = "Surface Area: " + strings.Title(record.Satellite.SurfaceArea)
	optionfive.Inline = true
	fields = append(fields, &optionfive)

	optionsix := discordgo.MessageEmbedField{}
	optionsix.Name = "6⃣"
	optionsix.Value = "Biosphere: " + strings.Title(record.Satellite.Biosphere)
	optionsix.Inline = true
	fields = append(fields, &optionsix)

	embed.Fields = fields

	var reactions []string
	reactions = append(reactions, "⬅")
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
	h.reactions.Watch(h.HandleSatelliteDetailsMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatelliteDetailsMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)

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
		h.SatellitePropertiesMenu(record, s, m)
		return
	}
	// DiscoveredBy
	if reaction == "1⃣" {
		h.SetSatelliteDiscoveredByMenu(record, s, m)
		return
	}
	// SystemZone
	if reaction == "2⃣" {
		h.SetSatelliteSystemZoneMenu(record, s, m)
		return
	}
	// Atmosphere
	if reaction == "3⃣" {
		h.SetSatelliteAtmosphereMenu(record, s, m)
		return
	}
	// Gravity
	if reaction == "4⃣" {
		h.SetSatelliteGravityMenu(record, s, m)
		return
	}
	// SurfaceArea
	if reaction == "5⃣" {
		h.SetSatelliteSurfaceAreaMenu(record, s, m)
		return
	}
	// Biosphere
	if reaction == "6⃣" {
		h.SetSatelliteBiosphereMenu(record, s, m)
		return

	}

	h.reactions.Watch(h.HandleSatelliteDetailsMenu, messageID, channelID, userID, recordname, s)
	return
}


// Discovered By

func (h *InfoHandler) SetSatelliteDiscoveredByMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Discovered By :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	if record.Satellite.DiscoveredBy == "" {
		optionone := discordgo.MessageEmbedField{}
		optionone.Name = ":bulb:"
		optionone.Value = "Current Value: " + record.Satellite.DiscoveredBy
		optionone.Inline = false
		fields = append(fields, &optionone)
	}

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new name or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteDiscoveredBy, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteDiscoveredByReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteDiscoveredBy(recordname string, userID string, name string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid names
	if len(name) > 1024 {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided name was too long, " +
			"must be shorter than 1024 characters for discord embed compatibility. Returning to previous menu.")
		h.SetSatelliteDiscoveredByMenu(record, s, m)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Discovered By"
	embed.Description = "Confirm Discovered By Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Name"
	optionone.Value = name
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Discovered By menu"
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

	payload := recordname + "|#|" + name
	h.reactions.Watch(h.HandleSetSatelliteDiscoveredByConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteDiscoveredByReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteDiscoveredByReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteDiscoveredByConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetRecordImageURLConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	discoveredby := payload[1]

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

		record.Satellite.DiscoveredBy = discoveredby
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteDiscoveredByMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu
		h.SetSatelliteDiscoveredByMenu(record, s, m)
		return
	}
}


// System Zone

func (h *InfoHandler) SetSatelliteSystemZoneMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite System Zone :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\"" +
		"\nCurrently: " + record.Satellite.SystemZone
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Average"
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Low"
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "High"
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	optionfour := discordgo.MessageEmbedField{}
	optionfour.Name = "4⃣"
	optionfour.Value = "Clear"
	optionfour.Inline = true
	fields = append(fields, &optionfour)

	embed.Fields = fields

	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	reactions = append(reactions, "3⃣")
	reactions = append(reactions, "4⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleSatelliteSystemZoneMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatelliteSystemZoneMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)

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
		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// DiscoveredBy
	if reaction == "1⃣" {
		record.Satellite.SystemZone = "average"
	} else if reaction == "2⃣" {
		record.Satellite.SystemZone = "low"
	} else if reaction == "3⃣" {
		record.Satellite.SystemZone = "high"
	} else if reaction == "4⃣" {
		record.Satellite.SystemZone = ""
	} else {
		h.reactions.Watch(h.HandleSatelliteDetailsMenu, messageID, channelID, userID, recordname, s)
		return
	}
	err = h.infodb.SaveRecordToDB(record, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	h.SetSatelliteDetailsMenu(record, s, m)
	return
}


// Atmosphere

func (h *InfoHandler) SetSatelliteAtmosphereMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Atmosphere :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":bulb:"
	optionone.Value = "Current Value: " + record.Satellite.Atmosphere
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new value or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteAtmosphere, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteAtmosphereReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteAtmosphere(recordname string, userID string, atmosphere string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid names
	_, err = strconv.ParseFloat(atmosphere, 64)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided value was not a float," +
			"please provide a correct value.")
		r := &discordgo.MessageReaction{MessageID:messageID, ChannelID: channelID, UserID: userID}
		h.SetSatelliteAtmosphereMenu(record, s, r)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Atmosphere"
	embed.Description = "Confirm Atmosphere Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Atmosphere"
	optionone.Value = atmosphere
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Atmosphere menu"
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

	payload := recordname + "|#|" + atmosphere
	h.reactions.Watch(h.HandleSetSatelliteAtmosphereConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteAtmosphereReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		//h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteAtmosphereReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteAtmosphereConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetRecordImageURLConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	atmosphere := payload[1]

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

		record.Satellite.Atmosphere = atmosphere
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteAtmosphereMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu

		h.SetSatelliteAtmosphereMenu(record, s, m)
		return
	}
}


// Gravity

func (h *InfoHandler) SetSatelliteGravityMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Gravity :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":bulb:"
	optionone.Value = "Current Value: " + record.Satellite.Gravity
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new value or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteGravity, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteGravityReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteGravity(recordname string, userID string, gravity string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid names
	_, err = strconv.ParseFloat(gravity, 64)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided value was not a float," +
			"please provide a correct value.")
		r := &discordgo.MessageReaction{MessageID:messageID, ChannelID: channelID, UserID: userID}
		h.SetSatelliteGravityMenu(record, s, r)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Gravity"
	embed.Description = "Confirm Gravity Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Gravity"
	optionone.Value = gravity
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Gravity menu"
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

	payload := recordname + "|#|" + gravity
	h.reactions.Watch(h.HandleSetSatelliteGravityConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteGravityReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		//h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteGravityReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteGravityConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetSatelliteGravityConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	gravity := payload[1]

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

		record.Satellite.Gravity = gravity
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteGravityMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu

		h.SetSatelliteGravityMenu(record, s, m)
		return
	}
}


// Surface Area

func (h *InfoHandler) SetSatelliteSurfaceAreaMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Surface Area :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":bulb:"
	optionone.Value = "Current Value: " + record.Satellite.SurfaceArea
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new value or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteSurfaceArea, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteSurfaceAreaReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteSurfaceArea(recordname string, userID string, surfacearea string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid names
	_, err = strconv.ParseFloat(surfacearea, 64)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided value was not a float," +
			"please provide a correct value.")
		r := &discordgo.MessageReaction{MessageID:messageID, ChannelID: channelID, UserID: userID}
		h.SetSatelliteSurfaceAreaMenu(record, s, r)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Surface Area"
	embed.Description = "Confirm Gravity Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Surface Area"
	optionone.Value = surfacearea
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Surface Area menu"
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

	payload := recordname + "|#|" + surfacearea
	h.reactions.Watch(h.HandleSetSatelliteSurfaceAreaConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteSurfaceAreaReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		//h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteSurfaceAreaReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteSurfaceAreaConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID,
			"Error: HandleSetSatelliteSurfaceAreaConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	surfacearea := payload[1]

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

		record.Satellite.SurfaceArea = surfacearea
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteSurfaceAreaMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		h.SetSatelliteSurfaceAreaMenu(record, s, m)
		return
	}
}


// Biosphere

func (h *InfoHandler) SetSatelliteBiosphereMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite System Zone :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\"" +
		"\nCurrently: " + record.Satellite.Biosphere
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Forest"
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Ice"
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "Desert"
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	optionfour := discordgo.MessageEmbedField{}
	optionfour.Name = "4⃣"
	optionfour.Value = "Clear"
	optionfour.Inline = true
	fields = append(fields, &optionfour)

	embed.Fields = fields

	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	reactions = append(reactions, "3⃣")
	reactions = append(reactions, "4⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleSatelliteBiosphereMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatelliteBiosphereMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)

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
		h.SetSatelliteDetailsMenu(record, s, m)
		return
	}
	// DiscoveredBy
	if reaction == "1⃣" {
		record.Satellite.Biosphere = "forest"
	} else if reaction == "2⃣" {
		record.Satellite.Biosphere = "ice"
	} else if reaction == "3⃣" {
		record.Satellite.Biosphere = "desert"
	} else if reaction == "4⃣" {
		record.Satellite.Biosphere = ""
	} else {
		h.reactions.Watch(h.HandleSatelliteBiosphereMenu, messageID, channelID, userID, recordname, s)
		return
	}
	err = h.infodb.SaveRecordToDB(record, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	h.SetSatelliteDetailsMenu(record, s, m)
	return
}


// Elements

func (h *InfoHandler) SetSatelliteElementsMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Elements :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	info := discordgo.MessageEmbedField{}
	info.Name = ":bulb:"
	elementlist := strings.Join(record.Satellite.NotableElements, ",")
	info.Value = "Current Elements: " + strings.Title(elementlist)
	info.Inline = false
	fields = append(fields, &info)


	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Add an Element"
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Remove an Element "
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	embed.Fields = fields

	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleSatelliteElementsMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatelliteElementsMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)

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
		h.SatellitePropertiesMenu(record, s, m)
		return
	}
	// Add Element
	if reaction == "1⃣" {
		h.AddSatelliteElementMenu(record, s, m)
		return
	}
	// Remove Element
	if reaction == "2⃣" {
		h.RemoveSatelliteElementMenu(record, s, m)
		return
	}


	// We only get this far if a wrong value was received
	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}
	h.reactions.Watch(h.HandleSatelliteElementsMenu, messageID, channelID, userID, recordname, s)
	return
}


func (h *InfoHandler) AddSatelliteElementMenu( record InfoRecord, s *discordgo.Session, m interface{}) {

	// Please provide an element name or press <- to return to the Satellite Properties Menu
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

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Satellite Add Element :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\"\nSelect an element from the list below"
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":pencil:"
	optionone.Value = "Select a new element below or ⬅ to return to the main menu"
	optionone.Inline = false
	fields = append(fields, &optionone)

	var reactions []string
	reactions = append(reactions, "⬅")

	// Element is a bit of a misnomer here since we're referring to a resource
	// This can become confusing, so keep in mind what we're talking about
	// When we refer to planetary elements
	// We want to get a list of resource records, and present them as options to be added, unless they are
	// already in the resource list, in which case we skip adding it.

	resourcelist, err := h.infodb.GetAllInfoResourceRecords(*collection)

	for i, resource := range resourcelist {
		if !StringSliceContains(record.Satellite.NotableElements, resource.Name) {
			resourcefield := discordgo.MessageEmbedField{}
			escapedemoji := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")

			resourcefield.Name = escapedemoji
			resourcefield.Value = strings.Title(resource.Name)
			resourcefield.Inline = true
			fields = append(fields, &resourcefield)
			reactions = append(reactions, escapedemoji)
		}
	}

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
	h.reactions.Watch(h.HandleSetSatelliteAddElementReactions, messageID, channelID, userID, record.Name, s)
	return


}

func (h *InfoHandler) HandleSetSatelliteAddElementReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteElementsMenu(record, s, m)
		return
	} else {
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

		resourcelist, err := h.infodb.GetAllInfoResourceRecords(*collection)

		for i := 0; i < 25; i++ {
			resourceindicator := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")
			if reaction == resourceindicator {
				record.Satellite.NotableElements = AppendStringIfMissing(record.Satellite.NotableElements, resourcelist[i].Name)
				err = h.infodb.SaveRecordToDB(record, *collection)
				if err != nil {
					_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
					return
				}
				h.SetSatelliteElementsMenu(record, s, m)
				return
			}
		}
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteAddElementReactions, messageID, channelID, userID, recordname, s)
	return

}


func (h *InfoHandler) RemoveSatelliteElementMenu( record InfoRecord, s *discordgo.Session, m interface{}) {

	// Please provide an element name or press <- to return to the Satellite Properties Menu
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
	embed.Title = "Info System Editor - Satellite Remove Element :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\"\nSelect an element from the list below"
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":pencil:"
	optionone.Value = "Select an element to remove or ⬅ to return to the main menu"
	optionone.Inline = false
	fields = append(fields, &optionone)

	var reactions []string
	reactions = append(reactions, "⬅")

	// Element is a bit of a misnomer here since we're referring to a resource
	// This can become confusing, so keep in mind what we're talking about
	// When we refer to planetary elements

	for i, resource := range record.Satellite.NotableElements {
		resourcefield := discordgo.MessageEmbedField{}
		escapedemoji := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")
		resourcefield.Name = escapedemoji
		resourcefield.Value = strings.Title(resource)
		resourcefield.Inline = true
		fields = append(fields, &resourcefield)
		reactions = append(reactions, escapedemoji)
	}

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
	h.reactions.Watch(h.HandleSetSatelliteRemoveElementReactions, messageID, channelID, userID, record.Name, s)
	return


}

func (h *InfoHandler) HandleSetSatelliteRemoveElementReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteElementsMenu(record, s, m)
		return
	} else {
		for i := 0; i < 25; i++ {
			resourceindicator := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")
			if reaction == resourceindicator {

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

				record.Satellite.NotableElements = append(record.Satellite.NotableElements[:i], record.Satellite.NotableElements[i+1:]...)
				err = h.infodb.SaveRecordToDB(record, *collection)
				if err != nil {
					_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
					return
				}

				h.SetSatelliteElementsMenu(record, s, m)
				return
			}
		}
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteAddElementReactions, messageID, channelID, userID, recordname, s)
	return

}



// Territories

// Territories Menu
func (h *InfoHandler) SetSatelliteTerritoriesMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Details :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "1⃣"
	optionone.Value = "Terra Nullius: " + strings.Title(record.Satellite.TerraNullius)
	optionone.Inline = true
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = "2⃣"
	optiontwo.Value = "Territory Count: " + strings.Title(strconv.Itoa(record.Satellite.Territories))
	optiontwo.Inline = true
	fields = append(fields, &optiontwo)

	optionthree := discordgo.MessageEmbedField{}
	optionthree.Name = "3⃣"
	optionthree.Value = "Territories Claimed: " + strings.Title(strconv.Itoa(record.Satellite.TerritoriesClaimed))
	optionthree.Inline = true
	fields = append(fields, &optionthree)

	embed.Fields = fields

	var reactions []string
	reactions = append(reactions, "⬅")
	reactions = append(reactions, "1⃣")
	reactions = append(reactions, "2⃣")
	reactions = append(reactions, "3⃣")
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.reactions.Watch(h.HandleSatelliteTerritoriesMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatelliteTerritoriesMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)

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
		h.SatellitePropertiesMenu(record, s, m)
		return
	}
	// Terra Nullius
	if reaction == "1⃣" {
		h.SetSatelliteTerraNulliusMenu(record, s, m)
		return
	}
	// Territory Count
	if reaction == "2⃣" {
		h.SetSatelliteTerritoryCountMenu(record, s, m)
		return
	}
	// Territories Claimed
	if reaction == "3⃣" {
		//h.SetSatelliteTerritoriesClaimedMenu(record, s, m)
		return
	}

	h.reactions.Watch(h.HandleSatelliteTerritoriesMenu, messageID, channelID, userID, recordname, s)
	return
}

// Terra Nullius

func (h *InfoHandler) SetSatelliteTerraNulliusMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Terra Nullius :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":bulb:"
	optionone.Value = "Current Value: " + record.Satellite.TerraNullius
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new value or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteTerraNullius, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteTerraNulliusReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteTerraNullius(recordname string, userID string, terranullius string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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

	// Handle invalid names
	_, err = strconv.ParseFloat(terranullius, 64)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided value was not a float," +
			"please provide a correct value.")
		r := &discordgo.MessageReaction{MessageID:messageID, ChannelID: channelID, UserID: userID}
		h.SetSatelliteAtmosphereMenu(record, s, r)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Terra Nullius"
	embed.Description = "Confirm Terra Nullius Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ImageURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Terra Nullius"
	optionone.Value = terranullius
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Terra Nullius menu"
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

	payload := recordname + "|#|" + terranullius
	h.reactions.Watch(h.HandleSetSatelliteTerraNulliusConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteTerraNulliusReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		//h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SatellitePropertiesMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteTerraNulliusReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteTerraNulliusConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetSatelliteTerraNulliusConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	terranullius := payload[1]

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

		record.Satellite.TerraNullius = terranullius
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteTerraNulliusMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu

		h.SetSatelliteTerraNulliusMenu(record, s, m)
		return
	}
}

// Territory Count

func (h *InfoHandler) SetSatelliteTerritoryCountMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Territory Count :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":bulb:"
	optionone.Value = "Current Value: " + strconv.Itoa(record.Satellite.Territories)
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new value below or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteTerritoryCount, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteTerritoryCountReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteTerritoryCount(recordname string, userID string, territorycount string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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
	_, err = strconv.Atoi(territorycount)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID,
			"Provided value is invalid, please check your entry and try again:\n```" + territorycount+"```")
		r := discordgo.MessageReaction{ChannelID:channelID, MessageID: messageID, UserID: userID}
		h.SetSatelliteTerritoryCountMenu(record, s, r)
		time.Sleep(15*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Territory Count"
	embed.Description = "Confirm Color Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Territory Count"
	optionone.Value = territorycount
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Territory Count menu"
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

	payload := recordname + "|#|" + territorycount
	h.reactions.Watch(h.HandleSetSatelliteTerritoryCountConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteTerritoryCountReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteTerritoriesMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteTerritoryCountReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteTerritoryCountConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetRecordColorConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	territorycount := payload[1]

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
		territorycountint, err := strconv.Atoi(territorycount)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			h.SetRecordThumbnailURLMenu(record, s, m)
			return
		}

		record.Satellite.Territories = territorycountint
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteTerritoryCountMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu
		h.SetSatelliteTerritoryCountMenu(record, s, m)
		return
	}
}

// Territories Claimed

func (h *InfoHandler) SetSatelliteTerritoriesClaimedMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

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
	embed.Title = "Info System Editor - Satellite Territories Claimed :satellite:"
	embed.Description = "Currently Editing: \"**"+strings.Title(record.Name)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":bulb:"
	optionone.Value = "Current Value: " + strconv.Itoa(record.Satellite.TerritoriesClaimed)
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":pencil:"
	optiontwo.Value = "Enter a new value below or select ⬅ to return to the main menu"
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
	h.infocallback.Watch(h.HandleSetSatelliteTerritoriesClaimed, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteTerritoriesClaimedReactions, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSetSatelliteTerritoriesClaimed(recordname string, userID string, territoriesclaimed string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)

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
	_, err = strconv.Atoi(territoriesclaimed)
	if err != nil {
		errormessage, _ := s.ChannelMessageSend(channelID,
			"Provided value is invalid, please check your entry and try again:\n```" + territoriesclaimed+"```")
		r := discordgo.MessageReaction{ChannelID:channelID, MessageID: messageID, UserID: userID}
		h.SetSatelliteTerritoryCountMenu(record, s, r)
		time.Sleep(15*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm Satellite Territories Claimed"
	embed.Description = "Confirm Color Selection For: \"**"+strings.Title(recordname)+"**\""
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Territories Claimed Count"
	optionone.Value = territoriesclaimed
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Territories Claimed menu"
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

	payload := recordname + "|#|" + territoriesclaimed
	h.reactions.Watch(h.HandleSetSatelliteTerritoriesClaimedConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteTerritoriesClaimedReactions(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	if reaction == "⬅" {
		h.infocallback.UnWatch(channelID, messageID, userID)
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

		h.SetSatelliteTerritoriesMenu(record, s, m)
		return
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleSetSatelliteTerritoriesClaimedReactions, messageID, channelID, userID, recordname, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteTerritoriesClaimedConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetSatelliteTerritoriesClaimedConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	territoriesclaimed := payload[1]

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
		territorycountint, err := strconv.Atoi(territoriesclaimed)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			h.SetRecordThumbnailURLMenu(record, s, m)
			return
		}

		record.Satellite.TerritoriesClaimed = territorycountint
		err = h.infodb.SaveRecordToDB(record, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.SetSatelliteTerritoriesClaimedMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu
		h.SetSatelliteTerritoriesClaimedMenu(record, s, m)
		return
	}
}