package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

type MainHandler struct {

	db *DBHandler
	conf *mainConfig
	dg *discordgo.Session

}

func (h *MainHandler) Init() error {

	// Add our main handler
	h.dg.AddHandler(h.Read)

	// Create a callback Handler and add it to our Handler Queue
	callback_handler := CallbackHandler{dg: h.dg}
	h.dg.AddHandler(callback_handler.Read)

	// Create our RSS handler
	rss := RSSHandler{db: h.db, conf: h.conf, callback: &callback_handler, dg: h.dg}
	h.dg.AddHandler(rss.Read)



	// Open a websocket connection to Discord and begin listening.
	err := h.dg.Open()
	if err != nil {
		fmt.Println("Error Opening Connection: ", err)
		return err
	}

	err = h.PostInit(h.dg)

	if err != nil {
		fmt.Println("Error during Post-Init")
		return err
	}

	return nil
}


// Just some quick things to run after our websocket has been setup and opened

func (h *MainHandler) PostInit(dg *discordgo.Session) error {

	// Update our default playing status
	err := h.dg.UpdateStatus(0, h.conf.DUBotConfig.Playing)
	if err != nil {
		fmt.Println("error updating now playing,", err)
		return err
	}

	return nil
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (h *MainHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == cp + "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == cp + "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == cp + "help" {
		s.ChannelMessageSend(m.ChannelID, "http://imgfave.com/collection/307305/Reaction-GIFs-no")
	}

}

