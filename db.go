package main

import (

	"github.com/asdine/storm"

	"fmt"
)

type DBHandler struct {

	DB *storm.DB
	conf *mainConfig

}

func (db *DBHandler) FirstTimeSetup() error {

	var user User

	err := db.DB.One("ID", db.conf.DiscordConfig.AdminID, &user)
	if err != nil {
		println("Running first time db config")
		user.ID = db.conf.DiscordConfig.AdminID
		user.SetRole("owner")
		err := db.DB.Save(&user)
		if err != nil{
			fmt.Println("error saving owner")
			return err
		}
		if(user.Owner){
			fmt.Println("Database has been configured")
			err = db.DB.One("ID", db.conf.DiscordConfig.AdminID, &user)
			fmt.Println("Owner ID: " + user.ID)
			return nil
		}
	}

	return nil
}



func (db *DBHandler) GetUser(uid string) (decoded User, err error) {

	var user User

	err = db.DB.One("ID", uid, &user)

	return user, err
}

