package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
)

// UserHandler struct
type UserHandler struct {
	querylocker sync.RWMutex

	conf    *Config
	db      *DBHandler
	cp      string
	logchan chan string
}

// Init function
func (h *UserHandler) Init() {
	h.cp = h.conf.DUBotConfig.CP
}

// Read function
func (h *UserHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.Fields(m.Content)

	if len(message) < 1 {
		return
	}

	h.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	// We use this a bit, this is the author id formatted as a mention
	mention := m.Author.Mention()

	if message[0] == cp+"groups" {
		mentions := m.Mentions

		if len(mentions) == 0 {
			groups, err := h.GetGroups(user.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving groups: "+err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, h.FormatGroups(groups))
			return
		}

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}

		if len(message) == 2 {
			groups, err := h.GetGroups(mentions[0].ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving groups: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "groups" + "||" + err.Error()
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatGroups(groups))
			return
		}
	}

	return
}

// GetUser function
func (h *UserHandler) GetUser(userid string) (user User, err error) {

	// Make sure user is in the database before we pull it out!
	h.CheckUser(userid)

	db := h.db.rawdb.From("Users")
	err = db.One("ID", userid, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}


// GetUser function
func (h *UserHandler) UpdateUserRecord(user User) (err error) {

	h.RemoveUserFromDB(user)
	if err != nil {
		return err
	}
	h.AddUser(user)
	if err != nil {
		return err
	}

	return nil
}

// AddToDB function
func (h *UserHandler) AddUser(user User) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Users")
	err = db.Save(&user)
	return err
}

// RemoveUserFromDB function
func (h *UserHandler) RemoveUserFromDB(user User) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Users")
	err = db.DeleteStruct(&user)
	return err
}

// CheckUser function
// This will register a new user into the database
func (h *UserHandler) CheckUser(ID string) {

	db := h.db.rawdb.From("Users")

	var u User
	err := db.One("ID", ID, &u)
	if err != nil {
		//fmt.Println("Adding new user to DB: " + ID)

		user := User{ID: ID}
		user.Init()

		err := db.Save(&user)
		if err != nil {
			fmt.Println("Error inserting user into Database!")
			return
		}

		walletdb := db.From("Wallets")

		var wallet Wallet
		err = walletdb.One("Account", ID, &wallet)
		if err != nil {
			wallet := Wallet{Account: ID}
			wallet.AddBalance(h.conf.BankConfig.SeedUserWalletBalance)
			err = walletdb.Save(&wallet)
			if err != nil {
				fmt.Println("Error inserting new wallet into Database!")
				return
			}
		} else {
			//fmt.Println("Wallet already exists for user!")
			return
		}

	}
}

// GetGroups function
func (h *UserHandler) GetGroups(ID string) (groups []string, err error) {

	h.CheckUser(ID)
	user, err := h.GetUser(ID)
	if err != nil {
		return groups, err
	}
	if user.CheckRole("owner") {
		groups = append(groups, "owner")
	}
	if user.CheckRole("admin") {
		groups = append(groups, "admin")
	}
	if user.CheckRole("smoderator") {
		groups = append(groups, "smoderator")
	}
	if user.CheckRole("moderator") {
		groups = append(groups, "moderator")
	}
	if user.CheckRole("editor") {
		groups = append(groups, "editor")
	}
	if user.CheckRole("agora") {
		groups = append(groups, "agora")
	}
	if user.CheckRole("streamer") {
		groups = append(groups, "streamer")
	}
	if user.CheckRole("recruiter") {
		groups = append(groups, "recruiter")
	}
	if user.CheckRole("citizen") {
		groups = append(groups, "citizen")
	}

	return groups, nil
}

// FormatGroups function
func (h *UserHandler) FormatGroups(groups []string) (formatted string) {
	for i, group := range groups {
		if i == len(groups)-1 {
			formatted = formatted + group
		} else {
			formatted = formatted + group + ", "
		}

	}

	return formatted
}
