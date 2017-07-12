package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MusicHandler struct {
	youtube      *YoutubeInterface
	db           *DBHandler
	user         *UserHandler
	registry     *CommandRegistry
	wallet       *WalletHandler
	channel      *ChannelHandler
	conf         *Config
	interrupt    chan string
	voicechannel string
	broadcasting bool
	songlive     bool
	paused       bool
	isSetup      bool
	vc           *discordgo.VoiceConnection

	nowplaying          string
	currentsongdetails  chan string
	nowplayingurl       string
	currentplaylist     string
	currentselectedby   string
	currentsongid       string
	currentsongduration time.Duration
	nextbuffer          [][]byte

	lastblockpos int

	restoresong bool

	bufferlocker sync.RWMutex
	initialized  bool
}

// Initializes our Music Handler
func (h *MusicHandler) Init() {
	h.youtube = &YoutubeInterface{db: h.db, conf: h.conf}
	h.interrupt = make(chan string)
	h.currentsongdetails = make(chan string)
}

// Reads our commands and passes them to the appropriate handlers
func (h *MusicHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.initialized {
		go h.StateManager(s, m)
		h.initialized = true
	}

	if !SafeInput(s, m, h.conf) {
		return
	}

	musicroomid, err := h.channel.GetMusicRoomChannel()
	if err != nil {
		return
	}

	if m.ChannelID != musicroomid {
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve user record: "+err.Error())
		return
	}

	if !user.Citizen {
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)

	// We don't need this here because we're already filtering for music room
	//	if command != "chess" {
	//		return
	//	}

	if command == "add" {
		if len(payload) < 1 {
			s.ChannelMessageSend(musicroomid, "<add> requires an argument")
			return
		}
		h.HandleAdd(user.ID, payload, s, m)
		return
	}
	if command == "play" {
		h.HandlePlay(payload, s, m)
		return
	}
	if command == "stop" {
		h.interrupt <- "stop"
		return
	}
	if command == "pause" {
		h.interrupt <- "pause"
		return
	}
	if command == "repeat" {
		return
	}
	if command == "next" {
		h.PlayNext(s, m)
		return
	}
	if command == "list" {
		return
	}
	if command == "nowplaying" || command == "info" {
		s.ChannelMessageSend(musicroomid, h.CurrentSongStatus(s))
		return
	}
	if command == "set" {
		if len(payload) < 1 {
			s.ChannelMessageSend(musicroomid, "<set> requires an argument")
			return
		}
		if payload[0] == "voicechannel" {
			h.SetVoiceChannel(s, m)
			return
		}
	}
}

func (h *MusicHandler) HandlePlay(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if h.paused {
		h.interrupt <- "pause"
		return
	}

	var playlist string
	var videoid string

	for _, message := range payload {
		if strings.HasPrefix(message, "id:") {
			videoid = strings.TrimPrefix(message, "id:")
			retrievedid, err := h.youtube.GetVideoID(videoid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error parsing id: "+err.Error())
				return
			}
			videoid = retrievedid
		}
		if strings.HasPrefix(message, "playlist:") {
			playlist = strings.TrimPrefix(message, "playlist:")
		}
	}
	if playlist == "" || playlist == "nil" {
		playlist = "default"
	}

	videoid, err := h.RetrieveVideo(playlist, videoid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving video: "+err.Error())
		return
	}

	if !h.isSetup || !h.broadcasting {
		err = h.SetupPlayback(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error setting up playback: "+err.Error())
			return
		}
	}

	if h.songlive {
		if h.paused {
			h.interrupt <- "pause"
		}
		h.interrupt <- "skip"
	}
	if h.restoresong {
		go h.StartPlayback(h.currentsongid, h.currentplaylist, true, s, m)
	} else {
		err = h.PrepareBuffer(videoid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error buffering audio: "+err.Error())
			return
		}
		go h.StartPlayback(videoid, playlist, false, s, m)
	}
	time.Sleep(1 * time.Second)
	s.ChannelMessageSend(m.ChannelID, h.CurrentSongStatus(s))
	return
}

// Playback Setup Functions

// Downloads the given video and converts it for playback.
func (h *MusicHandler) RetrieveVideo(playlist string, videoid string) (retrievedid string, err error) {

	playlistdb, err := h.youtube.GetDB(playlist)
	if err != nil {
		return "", err
	}

	for i, video := range playlistdb {

		if videoid == "" {
			videoid = video.VideoID
		}
		if video.VideoID == videoid {
			_, err = os.Stat("tmp/" + playlistdb[i].VideoID + ".opus")

			if err != nil {

				if os.IsNotExist(err) {
					err = h.youtube.DownloadYoutube(playlistdb[i].VideoID)

					if err != nil {
						return "", err
					}

					err = ToOpus(playlistdb[i].VideoID + ".mp4")
					if err != nil {
						return "", err
					}

					err = os.Remove("tmp/" + playlistdb[i].VideoID + ".mp4")
					if err != nil {
						return "", err
					}

					return videoid, nil
				}

				return "", err
			}

			return videoid, nil
		}
	}

	return "", errors.New("No matching video found in selected playlist")
}

// Sets up a new session. This will disconnect an active session, as it should only be called when no active session exists.
func (h *MusicHandler) SetupPlayback(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	if h.broadcasting {
		h.interrupt <- "disconnect"
	}
	if h.paused {
		h.interrupt <- "pause"
		h.interrupt <- "disconnect"
	}

	audioroom, err := h.channel.channeldb.GetMusicAudio()
	if err != nil {
		return err
	}

	// Check to see if our voice channel is empty or not
	if IsVoiceChannelEmpty(s, audioroom, s.State.User.ID) {
		return errors.New("Channel is empty!")
	}
	h.voicechannel = audioroom

	if !h.broadcasting {
		err = h.ConnectVoice(h.voicechannel, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error setting up voice connection: "+err.Error())
			return
		}
	}

	h.isSetup = true
	return nil
}

// Get Channel/Guild
func (h *MusicHandler) GetGuildID(s *discordgo.Session, m *discordgo.MessageCreate) (guildid string, err error) {

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return "", err
	}
	// Find the guild for that channel.
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return "", err
	}
	return guild.ID, nil
}

// Connects our bot to a voice channel and returns an error if anything goes wrong
func (h *MusicHandler) ConnectVoice(audiochannel string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	guildid, err := h.GetGuildID(s, m)
	if err != nil {
		return err
	}

	// Join the voice channel
	vc, err := s.ChannelVoiceJoin(guildid, audiochannel, false, true)
	if err != nil {
		return err
	}
	h.voicechannel = vc.ChannelID
	time.Sleep(250 * time.Millisecond)

	vc.Speaking(true)
	h.broadcasting = true
	h.vc = vc
	return nil
}

// Disconects our voice channel connect
func (h *MusicHandler) DisconnectVoice() {

	if h.vc == nil {
		return
	}

	h.broadcasting = false
	time.Sleep(250 * time.Millisecond)
	h.vc.Speaking(false)
	h.vc.Disconnect()
	h.vc = nil
	h.isSetup = false
}

func (h *MusicHandler) SetNowPlaying(videoid string, playlist string) {
	vid, err := h.youtube.GetVideoObject(videoid)
	if err != nil {
		return
	}
	record, err := h.youtube.GetFromDB(videoid, playlist)
	if err != nil {
		return
	}
	h.nowplaying = vid.Title
	h.nowplayingurl = "https://www.youtube.com/watch?v=" + vid.ID
	h.currentplaylist = playlist
	h.currentselectedby = record.UserID
	h.currentsongid = record.VideoID
	h.currentsongduration = vid.Duration
}

func (h *MusicHandler) UnSetNowPlaying() {

	h.nowplaying = "Nothing currently playing"
	h.nowplayingurl = "nil"
	h.currentplaylist = "nil"
	h.currentselectedby = "nil"
	h.currentsongid = "nil"
}

func (h *MusicHandler) PrepareBuffer(videoid string) (err error) {

	h.bufferlocker.Lock()
	defer h.bufferlocker.Unlock()

	buffer, err := MakeAudioBuffer("tmp/" + videoid + ".opus")
	if err.Error() != "EOF" {
		return err
	}

	// Flush the next buffer
	for i, _ := range h.nextbuffer {
		h.nextbuffer[i] = nil
	}

	h.nextbuffer = make([][]byte, len(buffer))
	copy(h.nextbuffer, buffer)

	for i, _ := range buffer {
		buffer[i] = nil
	}

	return nil
}

// This is typically what we'll be calling when we start a new song.
// Sets up our playback loop (buffer unpacking) and catches signals from the MusicHandler.interrupt channel.
// This expects a video that has already been processed, so DO NOT call it before retrieving and verifying there are no errors.
func (h *MusicHandler) StartPlayback(videoid string, playlist string, restore bool, s *discordgo.Session, m *discordgo.MessageCreate) {

	if h.nextbuffer == nil {
		return
	}
	h.bufferlocker.Lock()
	currentbuffer := make([][]byte, len(h.nextbuffer))
	copy(currentbuffer, h.nextbuffer)
	//if !restore {
	// Flush the next buffer
	//for i, _ := range h.nextbuffer {
	//	h.nextbuffer[i] = nil
	//}
	//}
	h.bufferlocker.Unlock()
	// Do not start playback if our voicechannel is nil
	if h.vc == nil {
		return
	}

	h.songlive = true
	h.SetNowPlaying(videoid, playlist)
	position := 0
	if restore {
		position = h.lastblockpos
	}

	// We restore our position if necessary.
	// the starting position should only change if we've loaded the previously loaded song
	for i := position; i < len(currentbuffer); i++ {
		h.vc.OpusSend <- currentbuffer[i]
		select {
		case msg := <-h.interrupt:
			if msg == "status" {
				h.currentsongdetails <- strconv.Itoa(len(currentbuffer)) + " " + strconv.Itoa(i)
			}
			if msg == "stop" {
				h.songlive = false
				h.restoresong = true
				h.lastblockpos = i
				for y, _ := range currentbuffer {
					currentbuffer[y] = nil
				}
				s.ChannelMessageSend(m.ChannelID, "Playback Stopped")
				return
			}
			if msg == "disconnect" {
				h.DisconnectVoice()
				h.songlive = false
				h.restoresong = true
				h.lastblockpos = i
				for y, _ := range currentbuffer {
					currentbuffer[y] = nil
				}
				s.ChannelMessageSend(m.ChannelID, "Playback Stopped")
				return
			}
			if msg == "skip" {
				h.songlive = false
				h.restoresong = false
				// Flush the buffer before returning
				for y, _ := range currentbuffer {
					currentbuffer[y] = nil
				}
				return
			}
			if msg == "pause" {
				h.paused = true
				s.ChannelMessageSend(m.ChannelID, "Playback paused")
				for h.paused {
					pmsg := <-h.interrupt
					if pmsg == "pause" {
						h.paused = false
						s.ChannelMessageSend(m.ChannelID, "Playback resumed")
					}
					if pmsg == "status" {
						h.currentsongdetails <- strconv.Itoa(len(currentbuffer)) + " " + strconv.Itoa(i)
					}
					if pmsg == "stop" {
						h.songlive = false
						h.paused = false
						h.restoresong = true
						h.lastblockpos = i
						for y, _ := range currentbuffer {
							currentbuffer[y] = nil
						}
						s.ChannelMessageSend(m.ChannelID, "Playback Stopped")
						return
					}
					if pmsg == "disconnect" {
						h.DisconnectVoice()
						h.songlive = false
						h.paused = false
						h.restoresong = true
						h.lastblockpos = i
						for y, _ := range currentbuffer {
							currentbuffer[y] = nil
						}
						s.ChannelMessageSend(m.ChannelID, "Playback Stopped")
						return
					}
					if pmsg == "skip" {
						h.songlive = false
						h.paused = false
						h.restoresong = false
						// Flush the buffer before returning
						for y, _ := range currentbuffer {
							currentbuffer[y] = nil
						}
						return
					}
				}
			}
		default:

		}
	}

	h.bufferlocker.Lock()
	// Flush the next buffer
	for i, _ := range h.nextbuffer {
		h.nextbuffer[i] = nil
	}
	h.bufferlocker.Unlock()
	for y, _ := range currentbuffer {
		currentbuffer[y] = nil
	}
	// If we got here our song ended and we don't need to unset anything but the fact that we have no live song.
	h.songlive = false
	h.PlayNext(s, m)
	return
}

func (h *MusicHandler) CurrentSongStatus(s *discordgo.Session) (status string) {

	if !h.songlive {
		return "No song currently playing"
	}

	h.interrupt <- "status"
	currentdetails := <-h.currentsongdetails

	details := strings.Fields(currentdetails)

	songduration := h.currentsongduration.String()
	sizeofbuffer, _ := strconv.Atoi(details[0])
	posinbuffer, _ := strconv.Atoi(details[1])

	currenttimeint := (float64(posinbuffer) / float64(sizeofbuffer)) * h.currentsongduration.Seconds()

	currenttime := time.Duration(time.Duration(currenttimeint) * time.Second).String()

	var username string
	user, err := s.User(h.currentselectedby)
	if err != nil {
		username = "error retrieving user"
	}
	username = user.Username

	output := ":musical_note: Now Playing || \n"
	output = output + "```\n"
	output = output + "Title: " + h.nowplaying + "\n"
	output = output + "Current Playlist: " + h.currentplaylist + "\n"
	output = output + "Song ID: " + h.currentsongid + "\n"
	output = output + "Added by: " + username + "\n"
	output = output + "Current Time: " + currenttime + "/" + songduration + "\n"
	output = output + "URL: https://www.youtube.com/watch?v=" + h.currentsongid + "\n"
	output = output + "```\n"
	return output
}

func (h *MusicHandler) PlayNext(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Send our command to skip the currently playing song
	if h.vc == nil {
		return
	}

	playlistname := h.currentplaylist
	if playlistname == "nil" {
		playlistname = "default"
	}

	playlist, err := h.youtube.GetDB("default")

	var index int
	for i, record := range playlist {
		if h.currentsongid == record.VideoID {

			if i == len(playlist)-1 {
				index = 0
			} else {
				if len(playlist) == 1 {
					index = 0
				} else {
					index = i + 1
				}
			}
		}
	}

	videoid := playlist[index].VideoID

	err = h.PrepareBuffer(videoid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error buffering audio: "+err.Error())
		return
	}
	if h.songlive {
		if h.paused {
			h.interrupt <- "pause"
		}
		h.interrupt <- "skip"
	}
	go h.StartPlayback(videoid, playlistname, false, s, m)
	time.Sleep(1 * time.Second)
	s.ChannelMessageSend(m.ChannelID, h.CurrentSongStatus(s)) // Startup our playback in another routine
	return
}

func (h *MusicHandler) PlayShuffle(audioroom string, videoid string, s *discordgo.Session, m *discordgo.MessageCreate) {

}

// Adds the item to the given playlist using argument syntax "playlist: genre: "
func (h *MusicHandler) HandleAdd(userid string, payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	// If we didn't provide arguments, put it into the default playlist with no genre
	if len(payload) == 1 {
		err := h.youtube.AddToPlaylist(payload[0], userid, "default", "nil")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error : "+err.Error())
			return
		}
		videoid, err := h.youtube.GetVideoID(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error : "+err.Error())
			return
		}

		id, err := h.RetrieveVideo("default", videoid)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error : "+err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Selection "+id+" added to default playlist.")
		return
	}

	playlist := ""
	genre := ""
	for _, argument := range payload {
		if strings.HasPrefix(argument, "playlist:") {
			playlist = strings.TrimPrefix(argument, "playlist:")
		}
		if strings.HasPrefix(argument, "genre:") {
			genre = strings.TrimPrefix(argument, "genre:")
		}
	}
	if playlist == "" {
		playlist = "default"
	}

	err := h.youtube.AddToPlaylist(payload[0], userid, playlist, genre)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error : "+err.Error())
		return
	}

	videoid, err := h.youtube.GetVideoID(payload[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error : "+err.Error())
		return
	}

	id, err := h.RetrieveVideo(playlist, videoid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error : "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Selection "+id+" added to "+playlist+" playlist.")
	return

}

// Used for setting up the voice channel that we broadcast to.
func (h *MusicHandler) SetVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) {

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving channel: "+err.Error())
		return
	}

	// Find the guild for that channel.
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving guild: "+err.Error())
		return
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			h.channel.channeldb.SetMusicAudio(vs.ChannelID)
			h.voicechannel = vs.ChannelID
			s.ChannelMessageSend(m.ChannelID, "Audio Room Set")
			return
		}
	}
	s.ChannelMessageSend(m.ChannelID, "No valid voice channel found!")
	return
}

// Go routine launched when we start audio playback to make sure we don't play to an empty room
func (h *MusicHandler) StateManager(s *discordgo.Session, m *discordgo.MessageCreate) {

	BotID := s.State.User.ID
	for {
		time.Sleep(time.Second * 5)
		voicechannel, err := h.channel.channeldb.GetMusicAudio()
		if err == nil {
			// If we stop broadcasting, we want to catch that and return
			if !h.broadcasting {
				if !IsVoiceChannelEmpty(s, voicechannel, BotID) {
					nothing := []string{""}
					h.paused = false
					h.HandlePlay(nothing, s, m)
				}
			}
			if h.broadcasting {
				if IsVoiceChannelEmpty(s, voicechannel, BotID) {
					//h.interrupt <- "stop"
					musicchan, err := h.channel.channeldb.GetMusicRoom()
					if err == nil {
						s.ChannelMessageSend(musicchan,
							"If a bot talks in an empty channel and nobody is around to hear it, does it still make a sound?")
						if h.songlive {
							h.interrupt <- "disconnect"
						} else {
							h.DisconnectVoice()
						}
					}
				}
			}
		}
	}
}
