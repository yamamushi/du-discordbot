package main

import (
	"github.com/bwmarrin/discordgo"
	"container/list"
	"reflect"
	"fmt"
)

// This will come in handy later
/*
	if reflect.TypeOf(m) == reflect.TypeOf(discordgo.MessageReactionAdd{}) {
		fmt.Println("Add found")
		//h.reactions.UnWatch(reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String(),
		//	reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String())
	}

	if reflect.TypeOf(m) == reflect.TypeOf(discordgo.MessageReactionRemove{}) {
		fmt.Println("Remove found")
		//h.reactions.UnWatch(reflect.Indirect(reflect.ValueOf(m)).FieldByName("ChannelID").String(),
		//	reflect.Indirect(reflect.ValueOf(m)).FieldByName("MessageID").String())
	}
*/

type ReactionsHandler struct {
	WatchList list.List
	dg        *discordgo.Session
	logger    *Logger
	conf     *Config

}


// WatchReaction struct
type WatchReaction struct {
	Reaction      string
	ChannelID     string
	MessageID     string
	Handler       func(string, string, *discordgo.Session, interface{})
	Args          string
}

// AddHandler function
func (h *ReactionsHandler) AddHandler(i interface{}) {
	// Important to note here, that this will only run once
	// We don't want to leave the handler running stale
	h.dg.AddHandlerOnce(i)
}

func (h *ReactionsHandler) Create(Handler func(string, string, *discordgo.Session, interface{}),
	Reactions []string, TargetChannelID string, Output string, Args string,
	s *discordgo.Session) (err error){

		message, err := s.ChannelMessageSend(TargetChannelID, Output)
		if err != nil {
			return err
		}
		//fmt.Println("ID: " + message.ID)
		//fmt.Println("Channel: " + message.ChannelID)

		for _, reaction := range Reactions {
			err = s.MessageReactionAdd(message.ChannelID, message.ID, reaction)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		h.Watch(Handler, message.ID, TargetChannelID, Args, s)
		return nil
}

// Watch function
func (h *ReactionsHandler) Watch(Handler func(string, string, *discordgo.Session, interface{}),
	MessageID string, TargetChannelID string,  Args string, s *discordgo.Session) {

	item := WatchReaction{ChannelID: TargetChannelID, MessageID: MessageID, Handler: Handler, Args: Args}
	h.WatchList.PushBack(item)

}

// UnWatch function
func (h *ReactionsHandler) UnWatch(ChannelID string, MessageID string) {

	// Clear user element by iterating
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		messageid := reflect.Indirect(r).FieldByName("MessageID")

		if channel.String() == ChannelID && messageid.String() == MessageID {
			h.WatchList.Remove(e)
		}
	}
}


func (h *ReactionsHandler) ReadReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd){

	// Ignore all messages created by the bot itself
	if m.UserID == s.State.User.ID {
		return
	}

	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channelid := reflect.Indirect(r).FieldByName("ChannelID").String()
		messageid := reflect.Indirect(r).FieldByName("MessageID").String()

		if m.MessageID == messageid && m.ChannelID == channelid {
			// We get the handler interface from our "Handler" field
			handler := reflect.Indirect(r).FieldByName("Handler")

			// We get our argument list from the Args field
			reflectedarglist := reflect.Indirect(r).FieldByName("Args")
			arglist := reflectedarglist.String()

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 4)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(m.Emoji.Name)
			rargs[1] = reflect.ValueOf(arglist)
			rargs[2] = reflect.ValueOf(s)
			rargs[3] = reflect.ValueOf(*m)

			handler.Call(rargs)
		}
	}
}


func (h *ReactionsHandler) ReadReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove){
	// Ignore all messages created by the bot itself
	if m.UserID == s.State.User.ID {
		return
	}

	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channelid := reflect.Indirect(r).FieldByName("ChannelID").String()
		messageid := reflect.Indirect(r).FieldByName("MessageID").String()

		if m.MessageID == messageid && m.ChannelID == channelid {
			// We get the handler interface from our "Handler" field
			handler := reflect.Indirect(r).FieldByName("Handler")

			// We get our argument list from the Args field
			reflectedarglist := reflect.Indirect(r).FieldByName("Args")
			arglist := reflectedarglist.String()

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 4)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(m.Emoji.Name)
			rargs[1] = reflect.ValueOf(arglist)
			rargs[2] = reflect.ValueOf(s)
			rargs[3] = reflect.ValueOf(*m)

			handler.Call(rargs)
		}
	}
}