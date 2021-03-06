package main

import (
	"bytes"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/yuin/gopher-lua"
	"io"
	"os"
	"strings"
	"time"
)

// LuaHandler struct
type LuaHandler struct {
	conf     *Config
	user     *UserHandler
	db       *DBHandler
	registry *CommandRegistry
}

// Read function
func (h *LuaHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !SafeInput(s, m, h.conf) {
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)
	if command != "lua" {
		return
	}

	h.user.CheckUser(m.Author.ID)

	_, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}
	/*
		command = payload[0]
		payload = RemoveStringFromSlice(payload, command)
	*/

	if command == "lua" {
		h.ReadLua(payload, s, m)
		return
	}

}

// ReadLua function
func (h *LuaHandler) ReadLua(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Under Construction!")
		return
	}

	command, payload := SplitPayload(payload)

	if command == "help" {
		s.ChannelMessageSend(m.ChannelID, "https://github.com/yamamushi/du-discordbot#lua")
		return
	}
	if command == "load" {
		s.ChannelMessageSend(m.ChannelID, "Under Construction!")
		return
	}
	if command == "read" {
		h.RunReadLuaInput(s, m)
		return
	}
	if command == "run" {
		//h.RunLua(payload, s, m)
		return
	}
	if command == "show" {
		s.ChannelMessageSend(m.ChannelID, "Under Construction!")
		return
	}
	if command == "save" {
		s.ChannelMessageSend(m.ChannelID, "Under Construction!")
		return
	}

}

// RunReadLuaInput function
func (h *LuaHandler) RunReadLuaInput(s *discordgo.Session, m *discordgo.MessageCreate) {

	message := m.Content

	if !strings.HasPrefix(message, h.conf.DUBotConfig.CP+"lua read ```") {
		//fmt.Println(message)
		s.ChannelMessageSend(m.ChannelID, "Invalid input!")
		return
	}

	message = strings.TrimPrefix(message, h.conf.DUBotConfig.CP+"lua read ```")

	if !strings.HasSuffix(message, "```") {
		//fmt.Println(message)
		s.ChannelMessageSend(m.ChannelID, "Invalid input!")
		return
	}

	message = strings.TrimSuffix(message, "```")

	if strings.Contains(message, "loadfile") {
		s.ChannelMessageSend(m.ChannelID, "loadfile is disabled!")
		return
	}

	if strings.Contains(message, "dofile") {
		s.ChannelMessageSend(m.ChannelID, "dofile is disabled!")
		return
	}

	if strings.Contains(message, "while true") || strings.Contains(message, "while 1") {
		s.ChannelMessageSend(m.ChannelID, "Invalid input!")
		return
	}

	if strings.Contains(message, "until false") || strings.Contains(message, "until nil") || strings.Contains(message, "until true") {
		s.ChannelMessageSend(m.ChannelID, "Invalid input!")
		return
	}

	h.RunLua(message, s, m)
	return
}

// RunLua function
func (h *LuaHandler) RunLua(script string, s *discordgo.Session, m *discordgo.MessageCreate) {

	old := os.Stdout     // keep backup of the real stdout
	r, w, _ := os.Pipe() // Create a new os pipe
	os.Stdout = w        // Reassign stdout to our temporary pipe

	// Create our new lua state
	l := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer l.Close()

	// Create our context pattern with a timeout as set in the config file
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.conf.DUBotConfig.LuaTimeout)*time.Second)
	defer cancel()

	// set the context to our LuaState
	l.SetContext(ctx)

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.CoroutineLibName, lua.OpenCoroutine},
		{lua.MathLibName, lua.OpenMath},
	} {
		if err := l.CallByParam(lua.P{
			Fn:      l.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}
	}

	err := l.DoString(script)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not execute script: "+err.Error())
		return
	}

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC   // assigning our channel content to our result string

	s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Your script result is:\n"+out)
	return
}
