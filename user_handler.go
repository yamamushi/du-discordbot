package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strconv"
	"strings"
)

type UserHandler struct {
	conf *Config
	db *DBHandler
}

func (h *UserHandler) Init() {

}

func (h *UserHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP
	db := h.db.rawdb.From("Users")

	if m.Author.ID == s.State.User.ID {
		return
	}

	message := strings.Fields(m.Content)

	var u User
	err := db.One("ID", m.Author.ID, &u)

	if err != nil {
		fmt.Println("Adding new user to DB!")
		user := User{ID: m.Author.ID}
		user.Init()
		err := db.Save(&user)
		if err != nil {
			fmt.Println("Error inserting user into Database!")
			return
		}

		walletdb := db.From("Wallets")

		var wallet Wallet
		err = walletdb.One("ID", m.Author.ID, &wallet)
		if err != nil {
			wallet := Wallet{Account: m.Author.ID}
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

	var user User
	err = db.One("ID", m.Author.ID, &user)
	if err != nil{
		fmt.Println("Error finding user")
		return
	}

	if message[0] == cp + "balance" && len(message) < 2 {

		userdb := h.db.rawdb.From("Users")
		walletdb := userdb.From("Wallets")

		var wallet Wallet
		err = walletdb.One("Account", m.Author.ID, &wallet)
		if err != nil {
			fmt.Println("Could not retrieve wallet for user! " + m.Author.ID)
			return
		}

		mention := m.Author.Mention()
		s.ChannelMessageSend(m.ChannelID, mention + " Your current balance is: " + strconv.Itoa(wallet.Balance))
		return
	}

	if message[0] == cp + "balance" && len(message) > 1 {

		mentions := m.Mentions
		if len(mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Invalid User: " + message[1])
			return
		}

		userdb := h.db.rawdb.From("Users")
		walletdb := userdb.From("Wallets")

		var wallet Wallet
		err = walletdb.One("Account", mentions[0].ID, &wallet)
		if err != nil {
			fmt.Println("Could not retrieve wallet for user! " + m.Author.ID)

			return
		}

		s.ChannelMessageSend(m.ChannelID, "Current balance is: " + strconv.Itoa( wallet.Balance))
		return
	}


	if message[0] == cp + "addbalance" && u.Admin {

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

		var wallet Wallet
		err = db.One("Account", u.ID, &wallet)
		if err != nil {
			fmt.Println("Could not retrieve wallet for user! " + m.Author.ID)
			return
		}

		wallet.AddBalance(i)
		mention := m.Author.Mention()
		db.Update(&wallet)
		s.ChannelMessageSend(m.ChannelID, mention + " Added " + message[1] + " to wallet")
	}



	if message[0] == cp + "transfer" {

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
		mentionedUser := mentions[0]

		var wallet Wallet
		err := db.One("Account", u.ID, &wallet)
		if err != nil {
			fmt.Println ("Error retrieving sender wallet!")
			s.ChannelMessageSend(m.ChannelID, "Internal error, sorry!")
			return
		}

		var testReceiver Wallet
		var newWallet bool = false
		err = db.One("Account", mentionedUser.ID, &testReceiver)
		if err != nil {
			wallet := Wallet{Account: mentionedUser.ID}
			wallet.AddBalance(1000)
			err = db.Save(&wallet)
			if err != nil {
				fmt.Println("Error inserting new wallet into Database!")
				return
			}
			newWallet = true
		}


		i, err := strconv.Atoi(message[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid value specified!: " + message[1])
			return
		}

		var receiver Wallet
		err = db.One("Account", mentionedUser.ID, &receiver)
		err = wallet.SendBalance(&receiver, i)
		if err != nil{
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		if newWallet {
			db.Save(&receiver)
		} else {
			db.Update(&receiver)
		}
		db.Update(&wallet)

		mentionReceiver := mentions[0].Mention()
		mentionSender := m.Author.Mention()
		s.ChannelMessageSend(m.ChannelID, mentionReceiver + " received " + message[1] + " from " + mentionSender)

	}

}
