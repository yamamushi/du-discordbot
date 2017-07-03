package main

import (
	"github.com/asdine/storm"
	"fmt"
)

type DBHandler struct {

	rawdb   *storm.DB
	conf *Config

}

func (h *DBHandler) FirstTimeSetup() error {

	var user User
	user.ID = h.conf.DiscordConfig.AdminID
	user.Init()

	db := h.rawdb.From("Users")

	err := db.One("ID", h.conf.DiscordConfig.AdminID, &user)
	if err != nil {
		fmt.Println("Running first time db config")
		walletdb := db.From("Wallets")
		user.SetRole("owner")
		err := db.Save(&user)
		if err != nil {
			fmt.Println("error saving owner")
			return err
		}

		wallet := Wallet{Account: h.conf.DiscordConfig.AdminID, Balance: 10000}
		err = walletdb.Save(&wallet)
		if err != nil {
			fmt.Println("error saving wallet")
			return err
		}

		if(user.Owner) {
			err = db.One("ID", h.conf.DiscordConfig.AdminID, &user)
			if err != nil{
				fmt.Println("Could not retrieve data from the database, something went wrong!")
				return err
			}
			fmt.Println("Owner ID: " + user.ID)
			fmt.Println("Database has been configured")
			return nil
		}
	}
	return nil
}

func (h *DBHandler) Insert(object interface{}) error {

	err := h.rawdb.Save(object)
	if err != nil {
		fmt.Println("Could not insert object: ", err.Error())
		return err
	}

	return nil
}

func (h *DBHandler) Find(first string, second string, object interface{}) error {

	err := h.rawdb.One(first, second, object)
	if err != nil {
		return err
	}
	return nil
}

func (h *DBHandler) Update(object interface{}) error {
	err := h.rawdb.Update(object)
	if err != nil {
		return err
	}
	return nil
}

func (h *DBHandler) GetUser(uid string) (user User, err error) {

	userdb := h.rawdb.From("Users")
	err = userdb.One("ID", uid, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

