package main

import (
	"github.com/bwmarrin/discordgo"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type BankHandler struct {

	conf *Config
	db *DBHandler
	logger *Logger
	user *UserHandler
	com *CommandHandler
	callback *CallbackHandler
	bank *Bank

}


func (h *BankHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !SafeInput(s, m, h.conf){
		return
	}

	command, payload :=  CleanCommand(m.Content, h.conf)

	h.user.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		//fmt.Println("Error finding user")
		return
	}

	if command ==  "bank" && user.Owner {
		h.Prompt(s, m)
		return
	}
	if command ==  "addbalance" && user.Owner {
		h.AddBalance(payload, s, m)
		return
	}
	if CheckPermissions(command, m.ChannelID, &user, s, h.com) {

		// We use this a bit, this is the author id formatted as a mention
		mention := m.Author.Mention()

		if command == "balance" && len(payload) < 1 {

			balance, err := h.GetBalance(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving balance: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, mention+" Your current balance is: "+balance)
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
	}

}


func (h *BankHandler) GetWallet(ID string) (Wallet, error) {

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


func (h *BankHandler) SaveWallet(wallet Wallet) (err error) {

	db := h.db.rawdb.From("Users").From("Wallets")

	err = db.DeleteStruct(&wallet)
	err = db.Save(&wallet)
	if err != nil {
		fmt.Println ("Error saving wallet!")
		return err
	}
	return nil
}

func (h *BankHandler) Transfer(message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

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
	h.logger.LogBank(mentionReceiver + " received " + message[0] + " from " + mentionSender ,s)

}


func (h *BankHandler) AddBalance (message []string, s *discordgo.Session, m *discordgo.MessageCreate) {

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


// Pedantic, but the extra verification is necessary
func (h *BankHandler) GetBalance (ID string) (string, error) {

	h.user.CheckUser(ID)

	wallet, err := h.GetWallet(ID)
	if err != nil {
		fmt.Println("Could not retrieve wallet for user! ")
		return "", errors.New("Could not retrieve wallet for user!")
	}

	return strconv.Itoa(wallet.Balance), nil

}


func (h *BankHandler) Prompt(s *discordgo.Session, m *discordgo.MessageCreate){

	var payload string

	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil{
		fmt.Println("Error creating user channel for " + m.Author.ID + " " + m.Author.Username )
		return
	}
/*
	// An MessageEmbed stores data for message embeds.
	type MessageEmbed struct {
		URL         string                 `json:"url,omitempty"`
		Type        string                 `json:"type,omitempty"`
		Title       string                 `json:"title,omitempty"`
		Description string                 `json:"description,omitempty"`
		Timestamp   string                 `json:"timestamp,omitempty"`
		Color       int                    `json:"color,omitempty"`
		Footer      *MessageEmbedFooter    `json:"footer,omitempty"`
		Image       *MessageEmbedImage     `json:"image,omitempty"`
		Thumbnail   *MessageEmbedThumbnail `json:"thumbnail,omitempty"`
		Video       *MessageEmbedVideo     `json:"video,omitempty"`
		Provider    *MessageEmbedProvider  `json:"provider,omitempty"`
		Author      *MessageEmbedAuthor    `json:"author,omitempty"`
		Fields      []*MessageEmbedField   `json:"fields,omitempty"`
	}
*/
	prompt := new(discordgo.MessageEmbed)

	prompt.URL = h.conf.BankConfig.BankURL
	prompt.Type = ""
	prompt.Title = "Bank Menu"
	prompt.Description = "Welcome to " + h.conf.BankConfig.BankName + " " + m.Author.Mention() + "!"
	prompt.Timestamp = ""
	prompt.Color = ColorGold()

	footer := new(discordgo.MessageEmbedFooter)
	footer.Text = h.conf.BankConfig.BankMenuSlogan
//	footer.IconURL = h.conf.BankConfig.BankIconURL
	prompt.Footer = footer
/*
	image := new(discordgo.MessageEmbedImage)
	image.URL = h.conf.BankConfig.BankIconURL
	image.Height = 5
	image.Width = 5
	prompt.Image = image

*/
	thumbnail := new(discordgo.MessageEmbedThumbnail)
	thumbnail.URL = h.conf.BankConfig.BankIconURL
	thumbnail.Height = 10
	thumbnail.Width = 10
	prompt.Thumbnail = thumbnail
/*
	video := new(discordgo.MessageEmbedVideo)
	video.URL = ""
	video.Height = 10
	video.Width = 10
	prompt.Video = video

	provider := new(discordgo.MessageEmbedProvider)
	provider.URL = ""
	provider.Name = ""
	prompt.Provider = provider
*/
	author := new(discordgo.MessageEmbedAuthor)
	author.Name = h.conf.BankConfig.BankName
	author.URL = h.conf.BankConfig.BankURL
	author.IconURL = "https://discordapp.com/api/users/"+s.State.User.ID+"/avatars/"+s.State.User.Avatar+".jpg"
	prompt.Author = author


	fields := []*discordgo.MessageEmbedField{}


	depositfield := discordgo.MessageEmbedField{}
	depositfield.Name = "deposit"
	depositfield.Value = "Deposit to account"
	depositfield.Inline = true
	fields = append(fields, &depositfield)

	withdrawfield := discordgo.MessageEmbedField{}
	withdrawfield.Name = "withdraw"
	withdrawfield.Value = "Withdraw from account"
	withdrawfield.Inline = true
	fields = append(fields, &withdrawfield)

	balancefield := discordgo.MessageEmbedField{}
	balancefield.Name = "balance"
	balancefield.Value = "Display account balance"
	balancefield.Inline = true
	fields = append(fields, &balancefield)

	transferfield := discordgo.MessageEmbedField{}
	transferfield.Name = "transfer"
	transferfield.Value = "Transfer to account"
	transferfield.Inline = true
	fields = append(fields, &transferfield)

	loansfield := discordgo.MessageEmbedField{}
	loansfield.Name = "loans"
	loansfield.Value = "Loan Management Menu"
	loansfield.Inline = true
	fields = append(fields, &loansfield)

	redeemfield := discordgo.MessageEmbedField{}
	redeemfield.Name = "redeem"
	redeemfield.Value = "Credit Redemption Menu"
	redeemfield.Inline = true
	fields = append(fields, &redeemfield)



	prompt.Fields = fields


	s.ChannelMessageSendEmbed(channel.ID, prompt)

	payload = channel.ID
	h.callback.Watch( h.ReadPrompt, GetUUID(), payload, s, m)
}




func (h *BankHandler) ReadPrompt(channelid string, s *discordgo.Session, m *discordgo.MessageCreate){

	/*
		Going to leave this check out for now, as anyone interacting with the bank through the public channels
		should realize what they're doing...

	if m.ChannelID != channelid {
		fmt.Println("channels don't match! " + m.ChannelID + " " + channelid)
		return
	}
	*/

	if !strings.HasPrefix(m.Content, h.conf.DUBotConfig.CP){
		s.ChannelMessageSend(channelid, "Invalid Command Received :: Banking Terminal Closed")
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)

	if command == "deposit" {
		h.ReadDeposit(payload, channelid, s, m)
		return
	}
	if command == "init" {
		h.InitBank(channelid, s, m)
		return
	}
	s.ChannelMessageSend(channelid, "Invalid Command Received :: Banking Terminal Closed")
	return

}



func (h *BankHandler) InitBank(channelid string, s *discordgo.Session, m *discordgo.MessageCreate){
	user, err := h.user.GetUser(m.Author.ID)
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	if !user.Owner {
		return
	}
	err = h.bank.CreateBank()
	if err != nil{
		s.ChannelMessageSend(channelid, err.Error())
		return
	}
	h.logger.LogBank("Bank Has Been Initialized", s)
	s.ChannelMessageSend(channelid, "Bank Has Been Initialized")
}




func (h *BankHandler) ReadDeposit (payload []string, channelid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.bank.BankInitialized() {
		h.logger.LogBank("Bank needs to be initialized!", s)
		s.ChannelMessageSend(channelid, "Owner has not initialized the bank yet and has been notified")
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(channelid, "<deposit> expects a value!")
		return
	}

	amount, err := strconv.Atoi(payload[0])
	if err != nil {
		s.ChannelMessageSend(channelid, "Invalid value supplied: "+payload[0])
		return
	}

	if len(payload) == 1{

		account, err := h.bank.GetAccountForUser(m.Author.ID)
		if err != nil{
			s.ChannelMessageSend(channelid, "Could not retrieve Bank Account: " + err.Error())
			return
		}

		wallet, err := h.GetWallet(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(channelid, "Could not retrieve wallet: " + err.Error())
			return
		}

		err = h.Deposit(amount, account.UserID, wallet)
		if err != nil{
			s.ChannelMessageSend(channelid, err.Error())
			return
		}

		h.logger.LogBank(m.Author.Mention() + " Deposited " + payload[0] + " into account " + account.ID, s)
		s.ChannelMessageSend(channelid, "Deposited " + payload[0] + " into account " + account.ID)
		return
	}
}



func (h *BankHandler) Deposit (amount int, userid string, wallet Wallet) (err error){

	account, err := h.bank.GetAccountForUser(userid)
	if err != nil {
		return err
	}

	if amount < 0 {
		return errors.New("Cannot send a negative number")
	}

	if amount > wallet.Balance {
		return errors.New("Insufficient funds in wallet")
	}

	wallet.Balance = wallet.Balance - amount
	account.Balance = account.Balance + amount

	h.SaveWallet(wallet)
	h.bank.SaveUserAccount(account)

	return nil

}

func (h *BankHandler) TransferToAccount (amount int, fromAccountID string, toAccountID string) (err error){

	return nil
}

func (h *BankHandler) Withdraw (amount int, accountID string, wallet Wallet) (err error){


	return nil
}


func (h *BankHandler) LoanRequest (amount int, accountID string) (response bool){

	return false

}