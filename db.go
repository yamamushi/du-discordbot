package main

import (

	"github.com/asdine/storm"
	"fmt"
)

type DBHandler struct {

	DB *storm.DB
	conf *mainConfig

}

func (h *DBHandler) FirstTimeSetup() error {

	var user User

	err := h.DB.One("ID", h.conf.DiscordConfig.AdminID, &user)
	if err != nil {
		println("Running first time db config")
		user.ID = h.conf.DiscordConfig.AdminID
		user.SetRole("owner")
		err := h.DB.Save(&user)
		if err != nil {
			fmt.Println("error saving owner")
			return err
		}
		if(user.Owner) {
			fmt.Println("Database has been configured")
			err = h.DB.One("ID", h.conf.DiscordConfig.AdminID, &user)
			fmt.Println("Owner ID: " + user.ID)
			return nil
		}
	}

	return nil

}


func (h *DBHandler) TransferOwner() error {

	return nil

}



func (h *DBHandler) GetUser(uid string) (decoded User, err error) {

	var user User

	err = h.DB.One("ID", uid, &user)

	return user, err
}

