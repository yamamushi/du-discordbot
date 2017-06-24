package main

import (

	"gopkg.in/oleiade/lane.v1"
	"github.com/bwmarrin/discordgo"
	"container/list"
	"reflect"
)

type CallbackHandler struct {

	Queue *lane.Queue
	WatchList list.List
	dg	*discordgo.Session

}

type WatchUser struct {

	User string
	ChannelID string

}


// We want to accept callback handlers
func (c *CallbackHandler) AddHandler(h interface{}) {

	// Important to note here, that this will only run once
	// We don't want to leave the handler running stale
	c.dg.AddHandlerOnce(h)

}


func (c *CallbackHandler) Watch(User string, ChannelID string) {

	item := WatchUser{User: User, ChannelID: ChannelID}
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

		if m.Author.ID == r_user.String() {
			c.dg.ChannelMessageSend(m.ChannelID, "Callback Success")
			c.UnWatch(m.Author.ID, m.ChannelID)
		}

	}
}