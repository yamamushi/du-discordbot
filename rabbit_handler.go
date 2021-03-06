package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/mgo.v2"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RabbitHandler struct {
	conf        *Config
	registry    *CommandRegistry
	db          *DBHandler
	userdb      *UserHandler
	globalstate *StateDB
	configdb    *ConfigDB

	backerdb *BackerInterface

	timeoutchan chan bool
	querylocker sync.RWMutex
	lastpost    time.Time
}

// Init function
func (h *RabbitHandler) Init() {
	h.RegisterCommands()
	h.timeoutchan = make(chan bool)
}

// RegisterCommands function
func (h *RabbitHandler) RegisterCommands() (err error) {
	h.registry.Register("rabbit", "Shhh", "check|count")
	return nil
}

// Read function
func (h *RabbitHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"rabbit") {
		if h.registry.CheckPermission("rabbit", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Citizen {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

func (h *RabbitHandler) CarrotFinder(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	if strings.ToLower(m.Content) != "🥕" {
		return
	}

	channellist, err := h.configdb.GetSettingList("rabbit-channel")
	if err != nil {
		return
	}

	found := false
	for _, channel := range channellist {
		if CleanChannel(channel) == m.ChannelID {
			found = true
		}
	}

	if !found {
		response, err := s.ChannelMessageSend(m.ChannelID, "Sorry, rabbits cannot be found in this channel!")
		if err == nil {
			time.Sleep(5 * time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}
		return
	}

	user, err := h.userdb.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if !user.Owner {
		response, err := s.ChannelMessageSend(m.ChannelID, "Thanks for the treat, but no rabbits for you! (You don't have permissions to lure a rabbit)")
		if err == nil {
			time.Sleep(5 * time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}
		return
	}

	lure, err := s.ChannelMessageSend(m.ChannelID, "You attempt to lure a rabbit (this is not guaranteed to work!).")
	if err == nil {
		time.Sleep(5 * time.Second)
		s.ChannelMessageDelete(m.ChannelID, lure.ID)
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
	h.Carrot(s, m)

	return
}

// Read function
func (h *RabbitHandler) Catch(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	if strings.ToLower(m.Content) != "catch" {
		return
	}

	channellist, err := h.configdb.GetSettingList("rabbit-channel")
	if err != nil {
		return
	}

	found := false
	for _, channel := range channellist {
		if CleanChannel(channel) == m.ChannelID {
			found = true
		}
	}

	if !found {
		response, err := s.ChannelMessageSend(m.ChannelID, "Sorry, rabbits cannot be found in this channel!")
		if err == nil {
			time.Sleep(5 * time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}
		return
	}

	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB(h.conf.DBConfig.MongoDB).C(h.conf.DBConfig.BackerRecordColumn)

	record, err := h.backerdb.GetRecordFromDB(m.Author.ID, *c)
	if err == nil {
		if record.PreAlpha == "true" || record.ATV == "true" || record.Alpha == "true" {
			response, err := s.ChannelMessageSend(m.ChannelID, "Sorry, you can only participate if you do not already have pre-alpha access.")
			if err == nil {
				time.Sleep(5 * time.Second)
				s.ChannelMessageDelete(m.ChannelID, response.ID)
				s.ChannelMessageDelete(m.ChannelID, m.ID)
				return
			}
		}
	}

	globalstate, err := h.globalstate.GetState()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if globalstate.RabbitLoose {

		user, err := h.db.GetUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		if user.RabbitWinner {
			response, err := s.ChannelMessageSend(m.ChannelID, "Sorry "+m.Author.Mention()+", you can only win once!")
			if err == nil {
				time.Sleep(5 * time.Second)
				s.ChannelMessageDelete(m.ChannelID, response.ID)
				s.ChannelMessageDelete(m.ChannelID, m.ID)
				return
			}
			return
		}

		user.RabbitCount = user.RabbitCount + 1
		h.userdb.UpdateUserRecord(user)

		rabbitWinCount, err := h.configdb.GetValue("rabbit-count")
		if err != nil {
			rabbitWinCount = int(h.conf.Rabbit.RabbitCount)
		}

		if user.RabbitCount >= rabbitWinCount {
			h.NotifyOwner(user.ID, s, m)
		}

		s.ChannelMessageSend(m.ChannelID, ":rabbit: "+m.Author.Mention()+" caught a rabbit!")

		globalstate.RabbitLoose = false

		err = h.globalstate.SetState(globalstate)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		return

	} else {
		response, err := s.ChannelMessageSend(m.ChannelID, "There are no rabbits in sight.")
		if err == nil {
			time.Sleep(5 * time.Second)
			s.ChannelMessageDelete(m.ChannelID, response.ID)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}
		return
	}

}

// ParseCommand function
func (h *RabbitHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	command, payload := SplitPayload(commandlist)

	if len(payload) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Command "+command+" expects an argument, see help for usage.")
		return
	}
	if payload[0] == "help" {
		h.HelpOutput(s, m)
		return
	}
	if payload[0] == "check" {
		_, commandpayload := SplitPayload(payload)
		h.CheckActive(commandpayload, s, m)
		return
	}
	if payload[0] == "count" {
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "count expects a user mention")
			return
		}
		h.GetCount(m.Mentions[0].ID, s, m)
		return
	}
	if payload[0] == "reward" {
		if len(m.Mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "reward expects a user mention")
			return
		}
		h.RewardUser(m.Mentions[0].ID, s, m)
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Unrecognized option: "+payload[0])
	return
}

func (h *RabbitHandler) HelpOutput(s *discordgo.Session, m *discordgo.MessageCreate) {

	output := ":rabbit2: Usage: \n```\n~rabbit check\n~rabbit count\n```\n"
	s.ChannelMessageSend(m.ChannelID, output)
	return

}

func (h *RabbitHandler) GetCount(userID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	user, err := h.userdb.GetUser(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	discorduser, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, discorduser.Username+" has "+strconv.Itoa(user.RabbitCount)+" rabbits in their inventory")
	return

}

func (h *RabbitHandler) CheckActive(payload []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	globalstate, err := h.globalstate.GetState()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if globalstate.RabbitLoose {
		s.ChannelMessageSend(m.ChannelID, "A rabbit is on the loose!")
		return
	} else {
		s.ChannelMessageSend(m.ChannelID, "There are no rabbits in sight.")
		return
	}

	return
}

func (h *RabbitHandler) RewardUser(userID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if !user.Owner {
		return // Silent return
	}

	rewardedUser, err := h.userdb.GetUser(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	rabbitWinCount, err := h.configdb.GetValue("rabbit-count")
	if err != nil {
		rabbitWinCount = int(h.conf.Rabbit.RabbitCount)
	}

	if rewardedUser.RabbitCount < rabbitWinCount {
		diff := rabbitWinCount - rewardedUser.RabbitCount
		if diff > 1 {
			s.ChannelMessageSend(m.ChannelID, "The selected user is not eligible for a prize yet, they need "+strconv.Itoa(diff)+" more rabbits to win.")
		} else {
			s.ChannelMessageSend(m.ChannelID, "The selected user is not eligible for a prize yet, they need "+strconv.Itoa(diff)+" more rabbit to win.")
		}
		return
	}

	rewardedUser.RabbitCount = 0
	rewardedUser.RabbitWinner = true

	err = h.userdb.UpdateUserRecord(rewardedUser)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	discorduser, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	unformatted, err := h.configdb.GetSetting("rabbit-channel")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error - Could not get rabbit-channel from configdb: "+err.Error())
		return
	}
	rabbitChannel := CleanChannel(unformatted)

	output := ":exclamation: " + discorduser.Mention() + " has caught " + strconv.Itoa(rabbitWinCount) + " rabbits and has won a prize!"
	s.ChannelMessageSend(rabbitChannel, output)
	return
}

func (h *RabbitHandler) NotifyOwner(userID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	ownerID := h.conf.DiscordConfig.AdminID

	userprivatechannel, err := s.UserChannelCreate(ownerID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error initializing backerauth.")
		return
	}

	user, err := s.User(userID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Please contact a staff member about this error - 41616")
		return
	}

	s.ChannelMessageSend(userprivatechannel.ID, user.Username+" has won, please verify their rabbit count and give them their prize!")
	return
}

func (h *RabbitHandler) Carrot(s *discordgo.Session, m *discordgo.MessageCreate) {
	db := h.db.rawdb.From("Users")
	var user User
	err := db.One("ID", m.Author.ID, &user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		return
	}

	if !user.Owner {
		return // Silent return
	}

	h.timeoutchan <- true
	//s.ChannelMessageSend(m.ChannelID, "Successfully forced the latest post from the recruitment queue")
	return
}

func (h *RabbitHandler) Release(s *discordgo.Session) {

	for {

		rabbitTimer, err := h.configdb.GetValue("rabbit-timer")
		if err != nil {
			rabbitTimer = int(h.conf.Rabbit.RabbitTimer)
		}

		randomOffset := rand.Intn(30)
		//time.Sleep((time.Duration(rabbitTimer)*time.Minute)+(time.Duration(randomOffset)*time.Minute))
		//time.Sleep((time.Duration(rabbitTimer)*time.Minute))

		select {
		case <-h.timeoutchan:
			break
		case <-time.After((time.Duration(rabbitTimer) * time.Minute) + (time.Duration(randomOffset) * time.Minute)):
			break
		}

		globalstate, err := h.globalstate.GetState()
		if err == nil {

			channellist, err := h.configdb.GetSettingList("rabbit-channel")
			if err == nil {

				rabbitRandom, err := h.configdb.GetValue("rabbit-random")
				if err != nil {
					rabbitRandom = int(h.conf.Rabbit.RabbitRandomWeight)
				}

				var channelrand int
				if len(channellist) == 1 {
					channelrand = 0
				} else {
					channelrand = rand.Intn(len(channellist))
				}

				if channelrand == len(channellist) {
					channelrand = channelrand - 1
				}
				rabbitChannel := CleanChannel(channellist[channelrand])

				randomresult := rand.Intn(100000)
				//fmt.Println(strconv.Itoa(randomresult))
				if randomresult < rabbitRandom {

					h.querylocker.Lock()
					globalstate.RabbitLoose = true

					err = h.globalstate.UpdateStateRecord(globalstate)
					if err == nil {
						s.ChannelMessageSend(rabbitChannel, ":rabbit: A rabbit hops into the room")
						h.querylocker.Unlock()

						rabbitExpire, err := h.configdb.GetValue("rabbit-expiration")
						if err != nil {
							rabbitExpire = int(h.conf.Rabbit.RabbitExpiration)
						}

						time.Sleep(time.Duration(rabbitExpire) * time.Minute)
						h.querylocker.Lock()
						globalstate, err := h.globalstate.GetState()
						if err == nil {
							if globalstate.RabbitLoose {
								globalstate.RabbitLoose = false

								h.globalstate.UpdateStateRecord(globalstate)
								s.ChannelMessageSend(rabbitChannel, ":rabbit2: The rabbit hops out of the room")
								h.querylocker.Unlock()
							} else {
								h.querylocker.Unlock()
							}
						} else {
							h.querylocker.Unlock()
						}
					} else {
						h.querylocker.Unlock()
					}
				}
			}
		}
	}
}
