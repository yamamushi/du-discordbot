package main

import (
	"errors"
	"fmt"
)

/*

As the name implies, the logic (if you can call it that) here controls interactions with the server bank.

The bank has a balance on hold that can be used for issuing loans.

Loan logic is also included here

*/

// Bank struct
type Bank struct {
	db   *DBHandler
	conf *Config
	user *UserHandler
}

// BankRecord struct
type BankRecord struct {
	ID           string `storm:"id"`
	Pin          string
	Balance      int
	LoansEnabled bool
}

// AccountRecord struct
type AccountRecord struct {
	ID         string `storm:"id"`
	Pin        string
	UserID     string `storm:"index"`
	Balance    int
	ActiveLoan bool `storm:"index"`
}

// Init function
func (h *Bank) Init() {

	if h.conf.BankConfig.Reset {
		err := h.ResetBank()
		if err != nil {
			fmt.Println("Error resetting bank: " + err.Error())
			return
		}
		return
	}

	_, err := h.GetMainBankAccount()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("!!!!!! If you are seeing this message, You need to run ~init from the bank prompt !!!!!!")
		fmt.Println("!!!!!! If you are seeing this message, You need to run ~init from the bank prompt !!!!!!")
		fmt.Println("!!!!!! If you are seeing this message, You need to run ~init from the bank prompt !!!!!!")
		fmt.Println("!!!!!! If you are seeing this message, You need to run ~init from the bank prompt !!!!!!")
		return
	}
}

// CreateBank function
func (h *Bank) CreateBank() (err error) {

	_, err = h.GetMainBankAccount()
	if err == nil {
		return errors.New("Main Bank Account already exists, it must be removed first! (See Readme for Bank Reset Information)")
	}

	// Loans are disabled for now!
	uuid, err := GetUUID()
	if err != nil{
		return errors.New("Fatal Error generating UUID")
	}
	mainrecord := BankRecord{ID: uuid, Pin: h.conf.BankConfig.Pin,
		Balance: h.conf.BankConfig.SeedWallet, LoansEnabled: false}

	db := h.db.rawdb.From("Bank")
	/* Let's see if this works without this delete first
	err = db.DeleteStruct(&account)
	if err != nil{
		return err
	}
	*/
	err = db.Save(&mainrecord)
	if err != nil {
		return err
	}

	return nil
}

// ResetBank function
func (h *Bank) ResetBank() (err error) {
	account, err := h.GetMainBankAccount()
	if err != nil {
		return err
	}

	db := h.db.rawdb.From("Bank")

	err = db.DeleteStruct(&account)
	if err != nil {
		return err
	}

	return h.CreateBank()
}

// SaveBank function
func (h *Bank) SaveBank() (err error) {
	account, err := h.GetMainBankAccount()
	if err != nil {
		return err
	}

	db := h.db.rawdb.From("Bank")
	/* Let's see if this works without this delete first
	err = db.DeleteStruct(&account)
	if err != nil{
		return err
	}
	*/
	err = db.Save(&account)
	if err != nil {
		return err
	}
	return nil
}

// BankInitialized function
func (h *Bank) BankInitialized() bool {

	_, err := h.GetMainBankAccount()
	if err != nil {
		return false
	}
	return true
}

// GetMainBankAccount function
func (h *Bank) GetMainBankAccount() (account BankRecord, err error) {

	db := h.db.rawdb.From("Bank")

	err = db.One("Pin", h.conf.BankConfig.Pin, &account)
	if err != nil {
		return account, err
	}

	return account, nil
}

// GetAccountForUser function
func (h *Bank) GetAccountForUser(userid string) (account AccountRecord, err error) {

	if !h.CheckUserAccount(userid) {
		h.CreateUserAccount(userid)
	}

	bankdb := h.db.rawdb.From("Bank")
	accountdb := bankdb.From("Accounts")

	record := AccountRecord{}
	err = accountdb.One("UserID", userid, &record)
	if err != nil {
		return record, err
	}
	return record, nil
}

// GetAccountByAccountID function
func (h *Bank) GetAccountByAccountID(accountid string) (account AccountRecord, err error) {

	bankdb := h.db.rawdb.From("Bank")
	accountdb := bankdb.From("Accounts")

	err = accountdb.One("ID", accountid, &account)
	if err != nil {
		return account, err
	}

	return account, nil
}

// CheckUserAccount function
func (h *Bank) CheckUserAccount(userid string) bool {

	bankdb := h.db.rawdb.From("Bank")
	accountdb := bankdb.From("Accounts")

	account := AccountRecord{}
	err := accountdb.One("UserID", userid, &account)

	if err != nil {
		return false
	}

	return true
}

// CreateUserAccount function
func (h *Bank) CreateUserAccount(userid string) (err error) {

	bankdb := h.db.rawdb.From("Bank")
	accountdb := bankdb.From("Accounts")

	account := AccountRecord{}
	err = accountdb.One("UserID", userid, &account)
	if err == nil {
		return errors.New("User Account Already Exists")
	}

	account.ID, err = GetUUID()
	if err != nil {
		return err
	}
	account.Pin = ""
	account.UserID = userid
	account.Balance = h.conf.BankConfig.SeedUserAccountBalance
	account.ActiveLoan = false

	err = h.SaveUserAccount(account)
	if err != nil {
		return err
	}
	return nil
}

// SaveUserAccount function
func (h *Bank) SaveUserAccount(account AccountRecord) (err error) {

	bankdb := h.db.rawdb.From("Bank")
	accountdb := bankdb.From("Accounts")

	if h.CheckUserAccount(account.UserID) {
		err = accountdb.DeleteStruct(&account)
		if err != nil {
			return err
		}
	}
	err = accountdb.Save(&account)
	if err != nil {
		return err
	}
	return nil
}
