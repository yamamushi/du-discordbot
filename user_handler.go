package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strconv"
	"strings"
	"errors"
)

type UserHandler struct {
	conf *Config
	db *DBHandler
	cp string
}

func (h *UserHandler) Init() {
	h.cp = h.conf.DUBotConfig.CP
}

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

	h.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println("Error finding user")
		return
	}

	// We use this a bit, this is the author id formatted as a mention
	mention := m.Author.Mention()

	if message[0] == cp + "balance" && len(message) < 2 {
		balance,err := h.GetBalance(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving balance: " + err.Error() )
			return
		}
		s.ChannelMessageSend(m.ChannelID, mention + " Your current balance is: " + balance)
		return
	}

	if message[0] == cp + "balance" && len(message) > 1 {
		mentions := m.Mentions
		if len(mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Invalid User: " + message[1])
			return
		}
		// Ignore bots
		if mentions[0].Bot {
			s.ChannelMessageSend(m.ChannelID, "Bots don't have money" )
			return
		}

		balance, err := h.GetBalance(mentions[0].ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving balance: " + err.Error() )
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Current balance is: " + balance)
		return
	}

	if message[0] == cp + "addbalance" && user.Admin {
		h.AddBalance(message, s, m)
		return
	}

	if message[0] == cp + "transfer" {
		h.Transfer(message, s, m)
		return
	}
}


func (h *UserHandler) CheckUser (ID string) {

	db := h.db.rawdb.From("Users")

	var u User
	err := db.One("ID", ID, &u)
	if err != nil {
		fmt.Println("Adding new user to DB: " + ID)

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
			wallet.AddBalance(1000)
			err = walletdb.Save(&wallet)
			if err != nil {
				fmt.Println("Error inserting new wallet into Database!")
				return
			}
		} else {
			fmt.Println("Wallet already exists for user!")
		}

	}
}


func (h *UserHandler) GetWallet(ID string) (Wallet, error) {

	db := h.db.rawdb.From("Users").From("Wallets")

	h.CheckUser(ID)

	var wallet Wallet
	err := db.One("Account", ID, &wallet)
	if err != nil {
		fmt.Println ("Error retrieving sender wallet!")
		return Wallet{}, err
	}
	return wallet, nil
}


func (h *UserHandler) Transfer(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	db := h.db.rawdb.From("Users").From("Wallets")
	if len(message) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Expected 2 arguments: transfer <amount> <recipient>")
		return
	}

	mentions := m.Mentions
	if len(mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Invalid Recipient: " + message[2])
		return
	}

	// Ignore bots
	if mentions[0].Bot {
		s.ChannelMessageSend(m.ChannelID, "Bots don't need money" )
		return
	}

	mentionedUser := mentions[0]

	sender, err := h.GetWallet(m.Author.ID)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, "Internal error, still in development!")
	}

	if m.Author.ID == mentionedUser.ID {
		s.ChannelMessageSend(m.ChannelID, "You cannot send money to yourself!")
		return
	}

	h.CheckUser(mentionedUser.ID)

	i, err := strconv.Atoi(message[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid value specified!: " + message[1])
		return
	}

	receiver,err := h.GetWallet(mentionedUser.ID)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, "Internal error, still in development!")
	}

	err = sender.SendBalance(&receiver, i)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	db.Update(&receiver)
	db.Update(&sender)

	mentionReceiver := mentions[0].Mention()
	mentionSender := m.Author.Mention()
	s.ChannelMessageSend(m.ChannelID, mentionReceiver + " received " + message[1] + " from " + mentionSender)

}


func (h *UserHandler) AddBalance (message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	db := h.db.rawdb.From("Users").From("Wallets")

	if len(message) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected Value!")
		return
	}

	i, err := strconv.Atoi(message[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid Input!")
		return
	}

	wallet, err := h.GetWallet(m.Author.ID)
	if err != nil {
		fmt.Println("Could not retrieve wallet for user! " + m.Author.ID)
		return
	}

	wallet.AddBalance(i)
	mention := m.Author.Mention()
	db.Update(&wallet)
	s.ChannelMessageSend(m.ChannelID, mention + " Added " + message[1] + " to wallet")
	return
}


// Pedantic, but the extra verification is necessary
func (h *UserHandler) GetBalance (ID string) (string, error) {

	h.CheckUser(ID)

	wallet, err := h.GetWallet(ID)
	if err != nil {
		fmt.Println("Could not retrieve wallet for user! ")
		return "", errors.New("Could not retrieve wallet for user!")
	}

	return strconv.Itoa(wallet.Balance), nil

}