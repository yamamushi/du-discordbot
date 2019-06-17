package main

import (
	"github.com/bwmarrin/discordgo"
	"html"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (h *InfoHandler) ViewSatelliteInfoMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	/*
	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	defer session.Close()
	*/

/*	flush := &discordgo.MessageEmbed{}
	flush.Title = strings.Title("Loading record...")
	_, err = s.ChannelMessageEditEmbed(channelID, messageID, flush)
	if err != nil {
		//fmt.Println(err.Error()) // We don't have to die here because this shouldn't be a fatal error (famous last words)
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	time.Sleep(2*time.Second)

 */
	collection, session, err := h.GetMongoCollecton()
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	defer session.Close()

	// We need to flush the old copy and get the latest one
	record, err = h.infodb.GetRecordFromDB(record.Name, *collection)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = strings.Title(record.Name)
	if record.Description != "" {
		embed.Description = record.Description
	}
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	if record.EditorID != "" {
		user, err := s.User(record.EditorID)
		if err == nil {
			embed.Footer = &discordgo.MessageEmbedFooter{Text: "Edited by: " + user.Username}
		}
	}
	if record.ImageURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{URL:record.ImageURL}
	}

	var fields []*discordgo.MessageEmbedField

	satellitetype := discordgo.MessageEmbedField{}
	if record.Satellite.SatelliteType == "planet" {
		satellitetype.Name = ":earth_asia:"
		satellitetype.Value = "**Planet**"
	} else if record.Satellite.SatelliteType == "moon" {
		satellitetype.Name = ":first_quarter_moon_with_face:"
		satellitetype.Value = "**Moon** of " + strings.Title(record.Satellite.ParentSatellite)
	}
	satellitetype.Inline = false
	fields = append(fields, &satellitetype )
	/*
	if len(record.Satellite.NotableElements) > 0 {
		notablelements := discordgo.MessageEmbedField{}
		notablelements.Name = "Notable Elements"
		notablelements.Value = strings.Title(strings.Join(record.Satellite.NotableElements, ", "))
		notablelements.Inline = true
		fields = append(fields, &notablelements)
	}

 */
	/*
	satellitecount := discordgo.MessageEmbedField{}
	satellitecount.Name = "Satellite Count"
	satellitecount.Value = strconv.Itoa(len(record.Satellite.Satellites))
	satellitecount.Inline = true
	fields = append(fields, &satellitecount)
	 */
	 /*
	if len(record.Satellite.Satellites) > 0 {
		satellites := discordgo.MessageEmbedField{}
		satellites.Name = "Satellites"
		satellites.Value = strings.Title(strings.Join(record.Satellite.Satellites, ", "))
		satellites.Inline = true
		fields = append(fields, &satellites)
	}
	  */

	var reactions []string
	// http://www.unicode.org/emoji/charts/full-emoji-list.html
	// https://www.rapidtables.com/convert/number/hex-to-decimal.html

	if len(record.Satellite.Satellites) > 0 {
		fullmoonemoji := html.UnescapeString("&#" + strconv.Itoa(127773) + ";")
		reactions = append(reactions, fullmoonemoji)
	}
	if record.Satellite.ParentSatellite != "" {
		newmoonemoji := html.UnescapeString("&#" + strconv.Itoa(127761) + ";")
		reactions = append(reactions, newmoonemoji)
	}

	sunbehindcloudemoji := html.UnescapeString("&#" + strconv.Itoa(127782) + ";")
	reactions = append(reactions, sunbehindcloudemoji)

	//checkeredflagemoji := html.UnescapeString("&#" + strconv.Itoa(127937) + ";")
	//reactions = append(reactions, checkeredflagemoji)

	raisedflagemoji := html.UnescapeString("&#" + strconv.Itoa(128681)+";")
	reactions = append(reactions, raisedflagemoji)

	//microscopeemoji := html.UnescapeString("&#" + strconv.Itoa(128300)+";")
	//reactions = append(reactions, microscopeemoji)

	pickaxeemoji := html.UnescapeString("&#" + strconv.Itoa(9935)+";")
	reactions = append(reactions, pickaxeemoji)

	if StringSliceContains(record.Satellite.UserList, userID) {
		wavinghandemoji := html.UnescapeString("&#" + strconv.Itoa(128075)+";")
		reactions = append(reactions, wavinghandemoji)
	} else {
		raisedhandemoji := html.UnescapeString("&#" + strconv.Itoa(129306)+";")
		reactions = append(reactions, raisedhandemoji)
	}


	stopsignemoji := html.UnescapeString("&#" + strconv.Itoa(128721)+";")
	reactions = append(reactions, stopsignemoji)

	embed.Fields = fields
	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}
	_, _ = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	h.reactions.Watch(h.HandleSatelliteInfoMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleSatelliteInfoMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {

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

	fullmoonemoji := html.UnescapeString("&#" + strconv.Itoa(127773) + ";")
	sunbehindcloudemoji := html.UnescapeString("&#" + strconv.Itoa(127782) + ";")
	wavinghandemoji := html.UnescapeString("&#" + strconv.Itoa(128075)+";")
	raisedhandemoji := html.UnescapeString("&#" + strconv.Itoa(129306)+";")
	raisedflagemoji := html.UnescapeString("&#" + strconv.Itoa(128681)+";")
	newmoonemoji := html.UnescapeString("&#" + strconv.Itoa(127761) + ";")
	pickaxeemoji := html.UnescapeString("&#" + strconv.Itoa(9935)+";")
	stopsignemoji := html.UnescapeString("&#" + strconv.Itoa(128721)+";")


	// Moons
	if reaction == newmoonemoji {
		parentsatellite, err := h.infodb.GetRecordFromDB(record.Satellite.ParentSatellite, *collection)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.ViewSatelliteInfoMenu(parentsatellite, s, m)
		return
	}
	// Moons
	if reaction == fullmoonemoji {
		h.ViewSatelliteMoonsMenu(record, s, m)
		return
	}
	// Atmospheric Details
	if reaction == sunbehindcloudemoji {
		h.ViewSatelliteDetailsMenu(record, s, m)
		return
	}
	// Unset Location
	if reaction == wavinghandemoji {
		h.HandleRemoveSatelliteLocation(record, s, m)
		return
	}
	// Set Location
	if reaction == raisedhandemoji {
		h.ViewSatelliteSetLocationMenu(record, s, m)
		return
	}
	// Territories
	if reaction == raisedflagemoji {
		h.ViewSatelliteTerritoriesMenu(record, s, m)
		return
	}

	if reaction == pickaxeemoji {
		h.ViewSatelliteElementsMenu(record, s, m)
		return
	}

	if reaction == stopsignemoji {
		_ = s.ChannelMessageDelete(channelID, messageID)
		s.ChannelMessageSend(channelID, "Info Session Closed")
		return
	}

	h.reactions.Watch(h.HandleSatelliteInfoMenu, messageID, channelID, userID, recordname, s)
	return
}


// Moons
func (h *InfoHandler) ViewSatelliteMoonsMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = strings.Title(record.Name)
	embed.Description = "Moons of " + strings.Title(record.Name)
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField
	var reactions []string
	reactions = append(reactions, "⬅")

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":last_quarter_moon_with_face: "
	optionone.Value = "Select a moon to view more information about it"
	optionone.Inline = false
	fields = append(fields, &optionone)

	for i, satellite := range record.Satellite.Satellites {
		satellitefield := discordgo.MessageEmbedField{}
		escapedemoji := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")
		satellitefield.Name = escapedemoji
		satellitefield.Value = strings.Title(satellite)
		satellitefield.Inline = true
		fields = append(fields, &satellitefield)
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
	h.reactions.Watch(h.HandleViewSatelliteMoonsMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleViewSatelliteMoonsMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {
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

		h.ViewSatelliteInfoMenu(record, s, m)
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


		for i := 0; i < 25; i++ {
			resourceindicator := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")
			if reaction == resourceindicator {
				moonrecord, err := h.infodb.GetRecordFromDB(record.Satellite.Satellites[i], *collection)
				if err != nil {
					_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
					return
				}
				h.ViewSatelliteInfoMenu(moonrecord, s, m)
				return
			}
		}
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleViewSatelliteMoonsMenu, messageID, channelID, userID, recordname, s)
	return
}


// Territories
func (h *InfoHandler) ViewSatelliteTerritoriesMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = strings.Title(record.Name)
	embed.Description = "Territories of " + strings.Title(record.Name)
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField
	var reactions []string
	reactions = append(reactions, "⬅")

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = strings.Title(record.Name)
	optionone.Value = "Territory Information for " + strings.Title(record.Name)
	optionone.Inline = false
	fields = append(fields, &optionone)

	terranullius := discordgo.MessageEmbedField{}
	terranullius.Name = "Terra Nullius"
	terranullius.Value = strings.Title(record.Satellite.TerraNullius)
	terranullius.Inline = false
	fields = append(fields, &terranullius)

	territories := discordgo.MessageEmbedField{}
	territories.Name = "Territories"
	territories.Value = strings.Title(strconv.Itoa(record.Satellite.Territories))
	territories.Inline = false
	fields = append(fields, &territories)

	territoriesclaimed := discordgo.MessageEmbedField{}
	territoriesclaimed.Name = "Territories Claimed"
	territoriesclaimed.Value = strings.Title(strconv.Itoa(record.Satellite.TerritoriesClaimed))
	territoriesclaimed.Inline = false
	fields = append(fields, &territoriesclaimed)

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
	h.reactions.Watch(h.HandleViewSatelliteTerritoriesMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleViewSatelliteTerritoriesMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	//messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	//userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()

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

	if reaction == "⬅" {
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	} else {
		// redundant for now but a placeholder for later
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	}

}

// Details
func (h *InfoHandler) ViewSatelliteDetailsMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = strings.Title(record.Name)
	embed.Description = ":white_sun_rain_cloud: Details of " + strings.Title(record.Name)
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField
	var reactions []string
	reactions = append(reactions, "⬅")

	if record.Satellite.DiscoveredBy != "" {
		discoveredby := discordgo.MessageEmbedField{}
		discoveredby.Name = "Discovered By"
		discoveredby.Value = strings.Title(record.Satellite.DiscoveredBy)
		discoveredby.Inline = false
		fields = append(fields, &discoveredby)
	}

	if record.Satellite.SystemZone != "" {
		systemzone := discordgo.MessageEmbedField{}
		systemzone.Name = "System Zone"
		systemzone.Value = strings.Title(record.Satellite.SystemZone)
		systemzone.Inline = true
		fields = append(fields, &systemzone)
	}

	if record.Satellite.Atmosphere != "" {
		atmosphere := discordgo.MessageEmbedField{}
		atmosphere.Name = "Atmosphere"
		atmosphere.Value = strings.Title(record.Satellite.Atmosphere)
		atmosphere.Inline = true
		fields = append(fields, &atmosphere)
	}

	if record.Satellite.Gravity != "" {
		gravity := discordgo.MessageEmbedField{}
		gravity.Name = "Gravity"
		gravity.Value = strings.Title(record.Satellite.Gravity)
		gravity.Inline = true
		fields = append(fields, &gravity)
	}

	if record.Satellite.SurfaceArea != "" {
		surfacearea := discordgo.MessageEmbedField{}
		surfacearea.Name = "Surface Area"
		surfacearea.Value = strings.Title(record.Satellite.SurfaceArea)
		surfacearea.Inline = true
		fields = append(fields, &surfacearea)
	}

	if record.Satellite.Biosphere != "" {
		biosphere := discordgo.MessageEmbedField{}
		biosphere.Name = "Biosphere"
		record.Satellite.Biosphere = strings.ToLower(record.Satellite.Biosphere)
		if record.Satellite.Biosphere == "ice" {
			coldfaceemoji := html.UnescapeString("&#" + strconv.Itoa(129398) + ";")
			biosphere.Value = coldfaceemoji + " Ice"
		} else if record.Satellite.Biosphere == "desert" {
			sunfaceemoji := html.UnescapeString("&#" + strconv.Itoa(127964) + ";")
			biosphere.Value = sunfaceemoji + " Desert"
		} else if record.Satellite.Biosphere == "forest" {
			treeemoji := html.UnescapeString("&#" + strconv.Itoa(127794) + ";")
			biosphere.Value = treeemoji + " Forest"
		} else if record.Satellite.Biosphere == "barren" {
			climbingemoji := html.UnescapeString("&#" + strconv.Itoa(129495) + ";")
			biosphere.Value = climbingemoji + " Barren"
		}
		biosphere.Inline = true
		fields = append(fields, &biosphere)
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
	h.reactions.Watch(h.HandleViewSatelliteTerritoriesMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleViewSatelliteDetailsMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

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

	if reaction == "⬅" {
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	} else {
		// redundant for now but a placeholder for later
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	}

}

// Satellite Elements
func (h *InfoHandler) ViewSatelliteElementsMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = strings.Title(record.Name)
	embed.Description = "Reources on " + strings.Title(record.Name)
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField
	var reactions []string
	reactions = append(reactions, "⬅")

	pickaxeemoji := html.UnescapeString("&#" + strconv.Itoa(9935)+";")

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = pickaxeemoji
	optionone.Value = "Select a resource to view more information about it (not currently implemented)"
	optionone.Inline = false
	fields = append(fields, &optionone)

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
	h.reactions.Watch(h.HandleViewSatelliteElementsMenu, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleViewSatelliteElementsMenu(reaction string, recordname string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	if reaction == "⬅" {
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

		h.ViewSatelliteInfoMenu(record, s, m)
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


		for i := 0; i < 25; i++ {
			resourceindicator := html.UnescapeString("&#" + strconv.Itoa(i+127462) + ";")
			if reaction == resourceindicator {
				_, err := h.infodb.GetRecordFromDB(record.Satellite.NotableElements[i], *collection)
				if err != nil {
					_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
					return
				}
				h.ViewSatelliteInfoMenu(record, s, m)
				// This should go to a resource rendering method which is not implemented quite yet
				// So instead we return to the main menu
				return
			}
		}
	}
	// If we got an invalid or unexpected reaction, ignore it and watch again
	h.reactions.Watch(h.HandleViewSatelliteMoonsMenu, messageID, channelID, userID, recordname, s)
	return
}


// User Location Registration
func (h *InfoHandler) ViewSatelliteSetLocationMenu(record InfoRecord, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := s.MessageReactionsRemoveAll(channelID, messageID)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{}
	embed.Title = strings.Title(record.Name)
	embed.Description = "Location Registration Menu - **" + strings.Title(record.Name) + "**"
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color

	var fields []*discordgo.MessageEmbedField
	var reactions []string
	reactions = append(reactions, "⬅")


	optionone := discordgo.MessageEmbedField{}
	optionone.Name = ":map:"
	optionone.Value = "Submit a coordinate below, or respond with \"confirm\" to set your location without an exact position"
	optionone.Inline = false
	fields = append(fields, &optionone)

	embed.Fields = fields

	for _, reaction := range reactions {
		_ = s.MessageReactionAdd(channelID, messageID, reaction)
	}

	_, err = s.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
		return
	}
	h.infocallback.Watch(h.HandleViewSatelliteSetLocationMenu, channelID, messageID, userID, record.Name, s)
	h.reactions.Watch(h.HandleSetSatelliteSetLocationNoPositionConfirm, messageID, channelID, userID, record.Name, s)
	return
}

func (h *InfoHandler) HandleViewSatelliteSetLocationMenu(recordname string, userID string, position string, s *discordgo.Session, m interface{}) {
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	// we have a userID here because we are passing a discordgo.MessageCreate interface which buries the userID under Author.ID
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)


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

	if position == "confirm" {
		err = h.SetUserLocation(userID, recordname, position)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		r := &discordgo.MessageReaction{MessageID:messageID,ChannelID:channelID,UserID:userID}
		h.ViewSatelliteInfoMenu(record, s, r)
		return
	}


	// Handle invalid names
	if !h.ValidatePosition(position) {
		errormessage, _ := s.ChannelMessageSend(channelID, "Provided location was invalid, returning to previous menu.")
		r := &discordgo.MessageReaction{MessageID:messageID,ChannelID:channelID,UserID:userID}
		h.ViewSatelliteInfoMenu(record, s, r)
		time.Sleep(10*time.Second)
		_ = s.ChannelMessageDelete(channelID, errormessage.ID)
		return
	}


	embed := &discordgo.MessageEmbed{}
	embed.Title = "Info System Editor - Confirm User Location"
	embed.Description = "Confirm Location"
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:record.ThumbnailURL}
	embed.Color = record.Color


	var fields []*discordgo.MessageEmbedField

	optionone := discordgo.MessageEmbedField{}
	optionone.Name = "Selected Position"
	optionone.Value = position
	optionone.Inline = false
	fields = append(fields, &optionone)

	optiontwo := discordgo.MessageEmbedField{}
	optiontwo.Name = ":question:"
	optiontwo.Value = "Select ✅ to confirm or ❎ to return to the Satellite Info menu"
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

	payload := recordname + "|#|" + position
	h.reactions.Watch(h.HandleSetSatelliteSetLocationConfirm, messageID, channelID, userID, payload, s)
	return

}

func (h *InfoHandler) HandleSetSatelliteSetLocationNoPositionConfirm(reaction string, recordname string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
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

	if reaction == "⬅" {
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	} else {
		h.reactions.Watch(h.HandleSetSatelliteSetLocationNoPositionConfirm, messageID, channelID, userID, record.Name, s)
		return
	}
}

func (h *InfoHandler) HandleSetSatelliteSetLocationConfirm(reaction string, args string, s *discordgo.Session, m interface{}) {

	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	payload := strings.Split(args, "|#|")
	if len(payload) < 2 || len(payload) > 2 {
		_, _ = s.ChannelMessageSend(channelID, "Error: HandleSetSatelliteSetLocationConfirm payload invalid size - " + strconv.Itoa(len(payload)))
		return
	}
	recordname := payload[0]
	position := payload[1]

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
		err = h.SetUserLocation(userID, record.Name, position)
		if err != nil {
			_, _ = s.ChannelMessageSend(channelID, "Error: " + err.Error())
			return
		}
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	} else {
		// 	if reaction == "❎"
		// Cancel and return to image url menu
		h.ViewSatelliteInfoMenu(record, s, m)
		return
	}
}

func (h *InfoHandler) HandleRemoveSatelliteLocation(record InfoRecord, s *discordgo.Session, m interface{}) {
	messageID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ID").String()
	channelID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String()
	userID := reflect.Indirect(reflect.ValueOf(m)).FieldByName("UserID").String()
	h.reactions.UnWatch(channelID, messageID, userID)
	h.infocallback.UnWatch(channelID, messageID, userID)

	err := h.SetUserLocation(userID, record.Name, "space")
	if err != nil {
		_, _ = s.ChannelMessageSend(channelID, "Error: "+err.Error())
		return
	}
	h.ViewSatelliteInfoMenu(record, s, m)
	return
}
