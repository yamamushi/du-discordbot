package main

import (
	"container/list"
	"github.com/bwmarrin/discordgo"
	"reflect"
)

// CallbackHandler struct
type CallbackHandler struct {
	WatchList list.List
	dg        *discordgo.Session
	logger    *Logger
}

// WatchUser struct
type WatchUser struct {
	User      string
	ChannelID string
	MessageID string
	Handler   func(string, *discordgo.Session, *discordgo.MessageCreate)
	Args      string
}

// AddHandler function
func (c *CallbackHandler) AddHandler(h interface{}) {
	// Important to note here, that this will only run once
	// We don't want to leave the handler running stale
	c.dg.AddHandlerOnce(h)
}

// Watch function
func (c *CallbackHandler) Watch(Handler func(string, *discordgo.Session, *discordgo.MessageCreate),
	MessageID string, Args string, s *discordgo.Session, m *discordgo.MessageCreate) {

	item := WatchUser{User: m.Author.ID, ChannelID: m.ChannelID, MessageID: MessageID, Handler: Handler, Args: Args}
	c.WatchList.PushBack(item)

}

// UnWatch function
func (c *CallbackHandler) UnWatch(User string, ChannelID string, MessageID string) {

	// Clear user element by iterating
	var next *list.Element
	for e := c.WatchList.Front(); e != nil; e = next {
		next = e.Next()

		r := reflect.ValueOf(e.Value)
		user := reflect.Indirect(r).FieldByName("User")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		messageid := reflect.Indirect(r).FieldByName("MessageID")

		if user.String() == User && channel.String() == ChannelID && messageid.String() == MessageID {
			c.WatchList.Remove(e)
		}
	}
}

// Read function
func (c *CallbackHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	var next *list.Element
	for e := c.WatchList.Front(); e != nil; e = next {

		r := reflect.ValueOf(e.Value)
		user := reflect.Indirect(r).FieldByName("User")
		channelid := reflect.Indirect(r).FieldByName("ChannelID")

		if m.Author.ID == user.String() && m.ChannelID == channelid.String() {

			// We get the handler interface from our "Handler" field
			handler := reflect.Indirect(r).FieldByName("Handler")

			// We get our argument list from the Args field
			arglist := reflect.Indirect(r).FieldByName("Args")
			command := arglist.String()

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 3)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(command)
			rargs[1] = reflect.ValueOf(s)
			rargs[2] = reflect.ValueOf(m)

			handler.Call(rargs)

			messageid := reflect.Indirect(r).FieldByName("MessageID").String()
			c.UnWatch(m.Author.ID, m.ChannelID, messageid)
		}
	}
}
