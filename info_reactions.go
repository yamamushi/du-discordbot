package main

import (
	"container/list"
	"github.com/bwmarrin/discordgo"
	"reflect"
	"sync"
	"time"
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

type InfoReactionsHandler struct {
	WatchList   list.List
	dg          *discordgo.Session
	logger      *Logger
	conf        *Config
	querylocker sync.RWMutex
	configdb    *ConfigDB
}

// WatchReaction struct
type WatchInfoReaction struct {
	Reaction  string
	ChannelID string
	MessageID string
	UserID    string
	Handler   func(string, string, *discordgo.Session, interface{})
	Created   time.Time
	Args      string
}

// AddHandler function
func (h *InfoReactionsHandler) AddHandler(i interface{}) {
	// Important to note here, that this will only run once
	// We don't want to leave the handler running stale
	h.dg.AddHandlerOnce(i)
}

func (h *InfoReactionsHandler) Create(Handler func(string, string, *discordgo.Session, interface{}),
	Reactions []string, TargetChannelID string, UserID string, Output string, Args string,
	s *discordgo.Session) (err error) {

	message, err := s.ChannelMessageSend(TargetChannelID, Output)
	if err != nil {
		return err
	}
	//fmt.Println("ID: " + message.ID)
	//fmt.Println("Channel: " + message.ChannelID)

	for _, reaction := range Reactions {
		_ = s.MessageReactionAdd(message.ChannelID, message.ID, reaction)
	}

	h.Watch(Handler, message.ID, TargetChannelID, UserID, Args, s)
	return nil
}

func (h *InfoReactionsHandler) CreateEmbed(Handler func(string, string, *discordgo.Session, interface{}),
	Reactions []string, TargetChannelID string, UserID string, Output *discordgo.MessageEmbed, Args string,
	s *discordgo.Session) (err error) {

	message, err := s.ChannelMessageSendEmbed(TargetChannelID, Output)
	if err != nil {
		return err
	}
	//fmt.Println("ID: " + message.ID)
	//fmt.Println("Channel: " + message.ChannelID)

	for _, reaction := range Reactions {
		_ = s.MessageReactionAdd(message.ChannelID, message.ID, reaction)
	}

	h.Watch(Handler, message.ID, TargetChannelID, UserID, Args, s)
	return nil
}

// Watch function
func (h *InfoReactionsHandler) Watch(Handler func(string, string, *discordgo.Session, interface{}),
	MessageID string, TargetChannelID string, UserID string, Args string, s *discordgo.Session) {

	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	item := WatchInfoReaction{ChannelID: TargetChannelID, MessageID: MessageID, UserID: UserID, Handler: Handler, Args: Args, Created: time.Now()}
	h.WatchList.PushBack(item)

}

func (h *InfoReactionsHandler) Cleaner() {
	for {
		time.Sleep(3 * time.Minute)
		expirationTime, err := h.configdb.GetValue("info-reactions-expiration")
		if err != nil {
			expirationTime = int(h.conf.Reactions.InfoReactionsExpiration)
		}

		h.querylocker.Lock()
		//fmt.Print("Locked")
		for e := h.WatchList.Front(); e != nil; e = e.Next() {
			r := reflect.ValueOf(e.Value)
			reaction := r.Interface().(WatchInfoReaction)
			if time.Now().After(reaction.Created.Add(time.Duration(expirationTime) * time.Minute)) {
				h.WatchList.Remove(e)
			}
		}
		h.querylocker.Unlock()
		//fmt.Print("Unlocked")
	}
}

// UnWatch function
func (h *InfoReactionsHandler) UnWatch(ChannelID string, MessageID string, UserID string) {

	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	// Clear user element by iterating
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		messageid := reflect.Indirect(r).FieldByName("MessageID")
		userid := reflect.Indirect(r).FieldByName("UserID")

		if channel.String() == ChannelID && messageid.String() == MessageID && userid.String() == UserID {
			h.WatchList.Remove(e)
		}
	}
}

func (h *InfoReactionsHandler) ReadReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {

	// Ignore all messages created by the bot itself
	if m.UserID == s.State.User.ID {
		return
	}

	//h.querylocker.Lock()
	//defer h.querylocker.Unlock()

	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channelid := reflect.Indirect(r).FieldByName("ChannelID").String()
		messageid := reflect.Indirect(r).FieldByName("MessageID").String()
		userid := reflect.Indirect(r).FieldByName("UserID").String()

		if m.MessageID == messageid && m.ChannelID == channelid && m.UserID == userid {
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

			h.WatchList.Remove(e)
			go handler.Call(rargs)
		}
	}
}

func (h *InfoReactionsHandler) ReadReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	// Ignore all messages created by the bot itself
	if m.UserID == s.State.User.ID {
		return
	}

	//h.querylocker.Lock()
	//h.querylocker.Unlock()

	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channelid := reflect.Indirect(r).FieldByName("ChannelID").String()
		messageid := reflect.Indirect(r).FieldByName("MessageID").String()
		userid := reflect.Indirect(r).FieldByName("UserID").String()

		if m.MessageID == messageid && m.ChannelID == channelid && m.UserID == userid {
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

			h.WatchList.Remove(e)
			go handler.Call(rargs)
		}
	}
}
