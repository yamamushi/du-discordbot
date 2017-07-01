package main

import "github.com/bwmarrin/discordgo"

type Logger struct {

	ch		*ChannelHandler

}

const (
	BOTLOG = iota
	PERMLOG
	BANKLOG
)

func (h *Logger) Init(ch *ChannelHandler){
	h.ch = ch
}

func (h *Logger) LogBot(message string, s *discordgo.Session){
	channelid, err := h.ch.GetBotLogChannel()
	if err != nil {
		return // Do nothing, we don't want to yell about no channel configured, just silently fail
	}

	s.ChannelMessageSend(channelid, message)
}


func (h *Logger) LogBank(message string, s *discordgo.Session){
	channelid, err := h.ch.GetBankLogChannel()
	if err != nil {
		return // Do nothing, we don't want to yell about no channel configured, just silently fail
	}

	s.ChannelMessageSend(channelid, message)
}

func (h *Logger) LogPerm(message string, s *discordgo.Session){
	channelid, err := h.ch.GetPermissionLogChannel()
	if err != nil {
		return // Do nothing, we don't want to yell about no channel configured, just silently fail
	}
	s.ChannelMessageSend(channelid, message)
}


func (h *Logger) Log(message string, s *discordgo.Session, level string){

	if level == "" {
		return
	}
	if level == "bot" {
		h.LogBot(message, s)
	}
	if level == "perm" || level == "permission"{
		h.LogPerm(message, s)
	}
	if level == "bank" {
		h.LogBank(message, s)
	}

}
