package main

import (

	"github.com/bwmarrin/discordgo"
	"container/list"
	"reflect"
	//"runtime"
	//"runtime"
)

type CallbackHandler struct {

	WatchList list.List
	dg	*discordgo.Session

}

type WatchUser struct {

	User string
	ChannelID string
	Handler func(string, *discordgo.Session, *discordgo.MessageCreate)
	Args string
}


// We want to accept callback handlers
func (c *CallbackHandler) AddHandler(h interface{}) {
	// Important to note here, that this will only run once
	// We don't want to leave the handler running stale
	c.dg.AddHandlerOnce(h)
}


func (c *CallbackHandler) Watch(User string, ChannelID string, Handler func(string, *discordgo.Session, *discordgo.MessageCreate),
	Args string, s *discordgo.Session, m *discordgo.MessageCreate) {

	item := WatchUser{User: User, ChannelID: ChannelID, Handler: Handler, Args: Args}
	c.WatchList.PushFront(item)

}


func (c *CallbackHandler) UnWatch(User string, ChannelID string) {

	// Clear user element by iterating
	var next *list.Element
	for e := c.WatchList.Front(); e != nil; e = next {

		r := reflect.ValueOf(e.Value)
		r_user := reflect.Indirect(r).FieldByName("User")
		r_channel := reflect.Indirect(r).FieldByName("ChannelID")

		if r_user.String() == User && r_channel.String() == ChannelID {
			c.WatchList.Remove(e)
		}
	}
}


func (c *CallbackHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate){

	var next *list.Element
	for e := c.WatchList.Front(); e != nil; e = next {

		r := reflect.ValueOf(e.Value)
		r_user := reflect.Indirect(r).FieldByName("User")
		r_channelid := reflect.Indirect(r).FieldByName("ChannelID")

		if m.Author.ID == r_user.String() && m.ChannelID == r_channelid.String() {
			//c.dg.ChannelMessageSend(m.ChannelID, "Callback Success")

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

			c.UnWatch(m.Author.ID, m.ChannelID)
		}
	}
}