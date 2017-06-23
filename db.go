package main

import (

	"github.com/asdine/storm"

	"fmt"
)

type DBHandler struct {

	DB *storm.DB
	conf *mainConfig

}

func (db *DBHandler) Configure() error {

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
	}

	err = db.DB.One("ID", db.conf.DiscordConfig.AdminID, &user)

	fmt.Println(user.ID)
	if(user.Owner){
		fmt.Println("Owner has been configured")
	}
	return err
}



func (db *DBHandler) GetUser(uid string) (decoded User, err error) {

	var user User

	err = db.DB.One("ID", uid, &user)

	return user, err
}

