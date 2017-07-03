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
	wallet *WalletHandler

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
	if len(payload) < 1 {
		if command ==  "bank" && user.Admin {
			h.Prompt(s, m)
			return
		}
	} else {
		if command ==  "bank" && user.Admin {
			payload = RemoveStringFromSlice(payload, command)
			channel, err := s.UserChannelCreate(m.Author.ID)
			if err != nil{
				fmt.Println("Error creating user channel for " + m.Author.ID + " " + m.Author.Username )
				return
			}
			payload[0] =  h.conf.DUBotConfig.CP+payload[0]
			var message string
			for _, content := range payload {
				message = message+content+" "
			}
			m.Content = message
			h.ReadPrompt(channel.ID, s, m)
			return
		}
	}
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
	useraccount, err := h.bank.GetAccountForUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not load account: " + err.Error())
		h.logger.LogBank(m.Author.Mention() + " Could not load account: " + err.Error(), s)
		return
	}

	prompt := new(discordgo.MessageEmbed)

	prompt.URL = h.conf.BankConfig.BankURL
	prompt.Type = ""
	prompt.Title = "Bank Menu"
	prompt.Description = "Welcome to " + h.conf.BankConfig.BankName + " " + m.Author.Mention() + "!"
	prompt.Description = prompt.Description + "\n" + "Account #" + useraccount.ID
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
	m.ChannelID = channel.ID
	h.callback.Watch( h.ReadPrompt, GetUUID(), payload, s, m)
}


// Bank Terminal Functions

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

	if command == "init" {
		if !h.bank.BankInitialized() {
			h.InitBank(channelid, s, m)
		}
		return
	}
	if command == "deposit" {
		h.ReadDeposit(payload, channelid, s, m)
		return
	}
	if command == "withdraw" {
		h.ReadWithdraw(payload, channelid, s, m)
		return
	}
	if command == "balance" {
		h.ReadBalance(payload, channelid, s, m)
		return
	}
	if command == "transfer" {
		h.ReadTransfer(payload, channelid, s, m)
		return
	}
	if command == "rewards" {
		h.ReadRewards(payload, channelid, s, m)
		return
	}
	if command == "loans" {
		h.ReadLoans(payload, channelid, s, m)
		return
	}


	fmt.Println(command)
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

		wallet, err := h.wallet.GetWallet(m.Author.ID)
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


func (h *BankHandler) ReadWithdraw (payload []string, channelid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.bank.BankInitialized() {
		h.logger.LogBank("Bank needs to be initialized!", s)
		s.ChannelMessageSend(channelid, "Owner has not initialized the bank yet and has been notified")
		return
	}

	if len(payload) < 1 {
		s.ChannelMessageSend(channelid, "<withdraw> expects a value!")
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

		wallet, err := h.wallet.GetWallet(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(channelid, "Could not retrieve wallet: " + err.Error())
			return
		}

		err = h.Withdraw(amount, account.UserID, wallet)
		if err != nil{
			s.ChannelMessageSend(channelid, err.Error())
			return
		}

		h.logger.LogBank(m.Author.Mention() + " Withdrew " + payload[0] + " from account " + account.ID, s)
		s.ChannelMessageSend(channelid, "Withdrew " + payload[0] + " from account " + account.ID)
		return
	}
}


func (h *BankHandler) ReadBalance (payload []string, channelid string, s *discordgo.Session, m *discordgo.MessageCreate){

	if !h.bank.BankInitialized() {
		h.logger.LogBank("Bank needs to be initialized!", s)
		s.ChannelMessageSend(channelid, "Owner has not initialized the bank yet and has been notified")
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(channelid, "Error Retrieving User")
		return
	}

	if len(payload) < 1 {

		account, err := h.bank.GetAccountForUser(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(channelid, "Could not find Bank Account: " + err.Error())
			return
		}
		balance := strconv.Itoa(account.Balance)
		s.ChannelMessageSend(channelid, "Your Account Bank Balance: " + balance + " credits")
		return
	}
	if len(payload) > 0 {

		if !user.Admin{
			s.ChannelMessageSend(channelid, "You do not have permission to view another Bank Account Balance.")
			if len(m.Mentions) > 0{
				h.logger.LogBank(m.Author.Mention() + " attempted to view bank balance for + " + m.Mentions[0].Mention(), s)
				return
			}
			h.logger.LogBank(m.Author.Mention() + " attempted to view bank balance for + " + payload[0], s)
			return
		}

		if len(m.Mentions) > 0 {
			account, err := h.bank.GetAccountForUser(m.Mentions[0].ID)
			if err != nil {
				s.ChannelMessageSend(channelid, "Error: Could not Bank Account Balance for " + m.Mentions[0].Mention()+" : "+err.Error())
				return
			}
			balance := strconv.Itoa(account.Balance)
			s.ChannelMessageSend(channelid, "Bank Account Balance for "+m.Mentions[0].Mention()+" : "+balance+" credits")
			return
		}

		// Verifies if a user account exists before proceeding, however we don't want to create one if it doesn't exist
		// Which GetAccountForUser will do.
		if !h.bank.CheckUserAccount(payload[0]){
			s.ChannelMessageSend(channelid, "Error: Could not Bank Account Balance for "+payload[0])
			return
		}

		//
		account, err := h.bank.GetAccountForUser(payload[0])
		if err != nil {
			s.ChannelMessageSend(channelid, "Error: Could not Bank Account Balance for "+m.Mentions[0].Mention()+" : "+err.Error())
			return
		}

		// Convert account balance to string and send t
		balance := strconv.Itoa(account.Balance)
		s.ChannelMessageSend(channelid, "Bank Account Balance for "+m.Mentions[0].Mention()+" : "+balance+" credits")
		return
	}
}


func (h *BankHandler) ReadTransfer (payload []string, channelid string, s *discordgo.Session, m *discordgo.MessageCreate){

	if !h.bank.BankInitialized() {
		h.logger.LogBank("Bank needs to be initialized!", s)
		s.ChannelMessageSend(channelid, "Owner has not initialized the bank yet and has been notified")
		return
	}

	if len(payload) < 2 {
		s.ChannelMessageSend(channelid, "<transfer> expects 2 values: transfer <value> <ToAccount#>")
		return
	}

	amount, err := strconv.Atoi(payload[0])
	if err != nil {
		s.ChannelMessageSend(channelid, "Invalid value supplied: "+payload[0])
		return
	}

	fromAccount, err := h.bank.GetAccountForUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(channelid, "Error getting sender user account: " + err.Error())
		return
	}

	toAccount, err := h.bank.GetAccountByAccountID(payload[1])
	if err != nil {
		s.ChannelMessageSend(channelid, "Error getting target user account: " + err.Error())
		return
	}

	err = h.TransferToAccount(amount, fromAccount.ID, toAccount.ID)
	if err != nil {
		s.ChannelMessageSend(channelid, "Could not transfer funds: " + err.Error())
		return
	}

	s.ChannelMessageSend(channelid, payload[0]+" credits transferred to " + payload[1])
	h.logger.LogBank(payload[0]+" credits transferred to " + payload[1],s)
	return

}


func (h *BankHandler) ReadRewards (payload []string, channelid string, s *discordgo.Session, m *discordgo.MessageCreate){
	s.ChannelMessageSend(channelid, "Not yet implemented.")
	return
}


func (h *BankHandler) ReadLoans (payload []string, channelid string, s *discordgo.Session, m *discordgo.MessageCreate){
	s.ChannelMessageSend(channelid, "Not yet implemented.")
	return
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

	h.wallet.SaveWallet(wallet)
	h.bank.SaveUserAccount(account)

	return nil

}


func (h *BankHandler) Withdraw (amount int, userid string, wallet Wallet) (err error){
	account, err := h.bank.GetAccountForUser(userid)
	if err != nil {
		return err
	}

	if amount < 0 {
		return errors.New("Cannot send a negative number")
	}

	if amount > account.Balance {
		return errors.New("Insufficient funds in account")
	}

	wallet.Balance = wallet.Balance + amount
	account.Balance = account.Balance - amount

	h.wallet.SaveWallet(wallet)
	h.bank.SaveUserAccount(account)
	return nil
}


func (h *BankHandler) TransferToAccount (amount int, fromAccountID string, toAccountID string) (err error){

	if amount < 1 {
		return errors.New("Cannot transfer a negative value!")
	}

	if fromAccountID == toAccountID {
		return errors.New("Invalid Bank Account ID Supplied!")
	}

	fromAccount, err := h.bank.GetAccountByAccountID(fromAccountID)
	if err != nil {
		return err
	}

	if amount > fromAccount.Balance {
		return errors.New("Insufficient Account Balance")
	}

	toAccount, err := h.bank.GetAccountByAccountID(toAccountID)
	if err != nil {
		return err
	}

	fromAccount.Balance = fromAccount.Balance - amount
	toAccount.Balance = toAccount.Balance + amount

	h.bank.SaveUserAccount(fromAccount)
	h.bank.SaveUserAccount(toAccount)

	return nil
}


func (h *BankHandler) LoanRequest (amount int, accountID string) (response bool){

	return false

}