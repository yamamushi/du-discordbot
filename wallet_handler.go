package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"errors"
)

type WalletHandler struct {

	db *DBHandler
	user *UserHandler
	logchan chan string
	conf *Config
}

func (h *WalletHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate){

	if !SafeInput(s, m, h.conf){
		return
	}

	command, payload :=  CleanCommand(m.Content, h.conf)

	if command != "wallet" {
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "<wallet> expects an argument")
		return
	}

	// Move our command up to the next element in the slice
	command = payload[0]
	payload = RemoveStringFromSlice(payload, command)

	h.user.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		//fmt.Println("Error finding user")
		return
	}

	if command ==  "addbalance" && user.Owner {
		h.AddBalance(payload, s, m)
		return
	}
	//if CheckPermissions(command, m.ChannelID, &user, s, h.com) {

	// We use this a bit, this is the author id formatted as a mention
	mention := m.Author.Mention()

	if command == "balance" && len(payload) < 1 {

		balance, err := h.GetBalance(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving balance: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, mention+" Your current wallet balance is: "+balance)
		return
	}

	if command == "balance" && len(payload) > 0 {
		mentions := m.Mentions
		if len(mentions) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Invalid User: " + payload[0])
			return
		}
		// Ignore bots
		if mentions[0].Bot {
			s.ChannelMessageSend(m.ChannelID, "Bots don't have money")
			return
		}

		balance, err := h.GetBalance(mentions[0].ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving balance: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Current balance is: "+balance)
		return
	}
	if command == "transfer" {
		h.Transfer(payload, s, m)
		return
	}
	//}
}

func (h *WalletHandler) GetWallet(ID string) (Wallet, error) {

	db := h.db.rawdb.From("Users").From("Wallets")

	h.user.CheckUser(ID)

	var wallet Wallet
	err := db.One("Account", ID, &wallet)
	if err != nil {
		fmt.Println ("Error retrieving sender wallet!")
		return Wallet{}, err
	}
	return wallet, nil
}


func (h *WalletHandler) SaveWallet(wallet Wallet) (err error) {

	db := h.db.rawdb.From("Users").From("Wallets")

	err = db.DeleteStruct(&wallet)
	err = db.Save(&wallet)
	if err != nil {
		fmt.Println ("Error saving wallet!")
		return err
	}
	return nil
}


func (h *WalletHandler) Transfer(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	db := h.db.rawdb.From("Users").From("Wallets")
	if len(message) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected 2 arguments: transfer <amount> <recipient>")
		return
	}

	mentions := m.Mentions
	if len(mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Invalid Recipient: " + message[1])
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

	h.user.CheckUser(mentionedUser.ID)

	amount, err := strconv.Atoi(message[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Invalid value specified!: " + message[0])
		return
	}

	receiver,err := h.GetWallet(mentionedUser.ID)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, "Internal error, still in development!")
	}

	err = sender.SendBalance(&receiver, amount)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	db.Update(&sender)
	db.Update(&receiver)


	mentionReceiver := mentions[0].Mention()
	mentionSender := m.Author.Mention()
	s.ChannelMessageSend(m.ChannelID, mentionReceiver + " received " + message[0] + " from " + mentionSender)
	h.logchan <- "Bank "+mentionReceiver + " received " + message[0] + " from " + mentionSender

}


func (h *WalletHandler) AddBalance (message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	db := h.db.rawdb.From("Users").From("Wallets")

	if len(message) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Expected Value!")
		return
	}

	i, err := strconv.Atoi(message[0])
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
	s.ChannelMessageSend(m.ChannelID, mention + " Added " + message[0] + " to wallet")
	return
}


func (h *WalletHandler) GetBalance (ID string) (string, error) {

	h.user.CheckUser(ID)

	wallet, err := h.GetWallet(ID)
	if err != nil {
		fmt.Println("Could not retrieve wallet for user! ")
		return "", errors.New("Could not retrieve wallet for user!")
	}

	return strconv.Itoa(wallet.Balance), nil

}
