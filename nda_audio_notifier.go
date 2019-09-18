package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/ogg"
	"os"
	"strings"
	"sync"
	"time"
)

type NDAAudioHandler struct {
	conf        *Config
	playing     bool
	querylocker sync.Mutex
	channels    []VoiceChannel
}

type VoiceChannel struct {
	ChannelID      string
	lastPlayedTime time.Time
}

func (h *NDAAudioHandler) Init() {

}

// This should be triggered when someone joins a voice channel
func (h *NDAAudioHandler) NDAWatch(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {

	if event.ChannelID == "" {
		return
	}
	if event.UserID == s.State.User.ID {
		return // Ignore ourselves
	}

	//fmt.Println("User Joined Voice Channel: " + event.ChannelID + " UserID: " + event.UserID)
	channel, err := s.Channel(event.ChannelID)
	if err != nil {
		fmt.Println("Error retrieving channel info: " + event.ChannelID + " " + err.Error())
		return
	}

	//fmt.Println("Channel Parent: " + channel.ParentID)
	category, err := s.Channel(channel.ParentID)
	if err != nil {
		fmt.Println("Error retrieving channel info: " + channel.ParentID + " " + err.Error())
		return
	}

	if !strings.Contains(category.Name, "NDA") {
		//fmt.Println("User joined a non-NDA voice channel")
		ndastatus, err := h.IsUserNDA(s, event.UserID)
		if err != nil {
			fmt.Println("Error retrieving NDA status: " + err.Error())
			return
		}

		if ndastatus {
			err = h.HandleNDAJoin(s, event.UserID, event.ChannelID)
			if err != nil {
				fmt.Println("Error with handling NDA user join: " + err.Error())
				return
			}
			return
		} else {
			err = h.HandleNonNDAJoin(s, event.UserID, event.ChannelID)
			if err != nil {
				fmt.Println("Error with handling NDA user join: " + err.Error())
				return
			}
			return
		}
	}
	return
}

func (h *NDAAudioHandler) HandleNonNDAJoin(s *discordgo.Session, userID string, channelID string) (err error) {
	if h.NDAUsersInJoinedChannel(s, userID, channelID) {
		err = h.PlayAudioFile(s, "./voice/channel_join.ogg", channelID, 2)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *NDAAudioHandler) HandleNDAJoin(s *discordgo.Session, userID string, channelID string) (err error) {
	if h.NonNDAUsersInJoinedChannel(s, userID, channelID) {
		err = h.PlayAudioFile(s, "./voice/nda_members.ogg", channelID, 2)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *NDAAudioHandler) WatchChannels(s *discordgo.Session) {
	for {
		time.Sleep(time.Minute * 20)
		for _, voicechannel := range h.channels {
			if h.NonNDAUsersInJoinedChannel(s, "", voicechannel.ChannelID) {
				err := h.PlayAudioFile(s, "./voice/nda_members.ogg", voicechannel.ChannelID, 2)
				if err != nil {
					fmt.Println("Error with audio playback: " + err.Error())
					return
				}
				return
			}
			return
		}
	}
}

func (h *NDAAudioHandler) TrackChannel(channelID string) {
	for _, voicechannel := range h.channels {
		if channelID == voicechannel.ChannelID {
			return
		}
	}
	then := time.Now().Add(time.Duration(-600) * time.Minute)
	voicechannel := VoiceChannel{ChannelID: channelID, lastPlayedTime: then}
	h.channels = append(h.channels, voicechannel)
}

func (h *NDAAudioHandler) PassesTimeCheck(channelID string) bool {
	for _, voicechannel := range h.channels {
		if channelID == voicechannel.ChannelID {
			elapsed := time.Since(voicechannel.lastPlayedTime)
			if elapsed.Minutes() < 10.0 {
				return false
			}
			return true
		}
	}
	h.TrackChannel(channelID)
	return true
}

func (h *NDAAudioHandler) NonNDAUsersInJoinedChannel(s *discordgo.Session, joinedUserID string, channelID string) bool {

	guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	if err != nil {
		fmt.Println(err.Error())
		return true
	}

	if guild.VoiceStates != nil {
		for _, voice := range guild.VoiceStates {
			// 514963150043611137 is the DU radio bot
			if voice.UserID != joinedUserID && voice.ChannelID == channelID && voice.UserID != "514963150043611137" {
				status, err := h.IsUserNDA(s, voice.UserID)
				if err != nil {
					fmt.Println(err.Error())
					return true
				}
				if !status {
					return true
				}
			}
		}
	}
	return false
}

func (h *NDAAudioHandler) NDAUsersInJoinedChannel(s *discordgo.Session, joinedUserID string, channelID string) bool {

	guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	if err != nil {
		return true
	}

	if guild.VoiceStates != nil {
		for _, voice := range guild.VoiceStates {
			if voice.UserID != joinedUserID && voice.ChannelID == channelID && voice.UserID != "514963150043611137" {
				status, err := h.IsUserNDA(s, voice.UserID)
				if err != nil {
					return true
				}
				if status {
					return true
				}
			}
		}
	}
	return false
}

func (h *NDAAudioHandler) IsUserNDA(s *discordgo.Session, userID string) (status bool, err error) {

	member, err := s.GuildMember(h.conf.DiscordConfig.GuildID, userID)
	if err != nil {
		return false, err
	}

	ndaid, err := getRoleIDByName(s, h.conf.DiscordConfig.GuildID, "Alpha Authorized")
	if err != nil {
		return false, err
	}

	for _, roleid := range member.Roles {
		if ndaid == roleid {
			return true, nil
		}
	}

	return false, nil
}

func (h *NDAAudioHandler) SetPlayingStatus(status bool) {
	h.playing = status
}

func (h *NDAAudioHandler) ResetLastPlayed(channelID string) {
	for i, voicechannel := range h.channels {
		if channelID == voicechannel.ChannelID {
			h.channels[i].lastPlayedTime = time.Now()
		}
	}
}

func (h *NDAAudioHandler) PlayAudioFile(s *discordgo.Session, path string, channelID string, pause int) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	h.playing = true
	defer h.SetPlayingStatus(false)

	if !h.PassesTimeCheck(channelID) {
		return
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(h.conf.DiscordConfig.GuildID, channelID, false, true)
	if err != nil {
		if _, ok := s.VoiceConnections[h.conf.DiscordConfig.GuildID]; ok {
			vc = s.VoiceConnections[h.conf.DiscordConfig.GuildID]
		} else {
			return err
		}
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(time.Duration(pause) * time.Second)

	// Start speaking.
	err = vc.Speaking(true)
	if err != nil {
		return err
	}

	// Send the buffer data.

	buffer, err := h.MakeAudioBuffer(path) // b has type []byte
	if err != nil {
		if err.Error() != "EOF" {
			fmt.Println("Could not create audio buffer: " + err.Error())
			_ = vc.Disconnect()
			return err
		}
	}

	for _, buf := range buffer {
		vc.OpusSend <- buf
	}
	h.ResetLastPlayed(channelID)

	// Stop speaking
	err = vc.Speaking(false)
	if err != nil {
		_ = vc.Disconnect()
		return err
	}

	// Sleep for a specificed amount of time before ending.
	time.Sleep(1 * time.Second)

	// Disconnect from the provided voice channel.
	_ = vc.Disconnect() // If we error here, what can we possibly complain about

	return nil
}

func (h *NDAAudioHandler) MakeAudioBuffer(path string) (output [][]byte, err error) {
	reader, err := os.Open(path)
	defer reader.Close()
	if err != nil {
		return output, err
	}
	oggdecoder := ogg.NewDecoder(reader)
	packetdecoder := ogg.NewPacketDecoder(oggdecoder)

	for {
		packet, _, err := packetdecoder.Decode()
		if err != nil {
			return output, err
		}
		output = append(output, packet)
	}
}
