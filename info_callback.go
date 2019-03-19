package main

import (
	"container/list"
	"github.com/bwmarrin/discordgo"
	"reflect"
)

// CallbackHandler struct
type InfoCallbackHandler struct {
	WatchList list.List
	dg        *discordgo.Session
	logger    *Logger
}

// WatchUser struct
type InfoWatchUser struct {
	UserID      string
	ChannelID string
	MessageID string
	Handler   func(string, string, string, *discordgo.Session, interface{})
	Args      string
}

// AddHandler function
func (c *InfoCallbackHandler) AddHandler(h interface{}) {
	// Important to note here, that this will only run once
	// We don't want to leave the handler running stale
	c.dg.AddHandlerOnce(h)
}

// Watch function
func (c *InfoCallbackHandler) Watch(Handler func(string, string, string, *discordgo.Session, interface{}),
	TargetChannelID string, MessageID string, UserID string, Args string, s *discordgo.Session) {

	item := InfoWatchUser{UserID: UserID, ChannelID: TargetChannelID, MessageID: MessageID, Handler: Handler, Args: Args}
	c.WatchList.PushBack(item)

}

// UnWatch function
func (c *InfoCallbackHandler) UnWatch(ChannelID string, MessageID string, UserID string) {

	// Clear user element by iterating
	for e := c.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		userid := reflect.Indirect(r).FieldByName("UserID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		messageid := reflect.Indirect(r).FieldByName("MessageID")

		if userid.String() == UserID && channel.String() == ChannelID && messageid.String() == MessageID {
			c.WatchList.Remove(e)
		}
	}
}

// Read function
func (c *InfoCallbackHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	for e := c.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		userid := reflect.Indirect(r).FieldByName("UserID")
		channelid := reflect.Indirect(r).FieldByName("ChannelID")
		messageid := reflect.Indirect(r).FieldByName("MessageID")

		if m.ChannelID == channelid.String() && m.Author.ID == userid.String() {

			_= s.ChannelMessageDelete(m.ChannelID, m.ID)

			// We get the handler interface from our "Handler" field
			handler := reflect.Indirect(r).FieldByName("Handler")

			// We get our argument list from the Args field
			reflectedarglist := reflect.Indirect(r).FieldByName("Args")
			arglist := reflectedarglist.String()

			m.ID = messageid.String()

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 5)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(arglist)
			rargs[1] = reflect.ValueOf(m.Author.ID)
			rargs[2] = reflect.ValueOf(m.Content)
			rargs[3] = reflect.ValueOf(s)
			rargs[4] = reflect.ValueOf(*m)

			go handler.Call(rargs)
			c.WatchList.Remove(e)
		}
	}
}
