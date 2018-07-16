package main

import (
	"github.com/kevinburke/go.uuid"
)

// GetUUID function
func GetUUID() (id string, err error) {

	formattedid := uuid.NewV4()
	//if err != nil {
	//	return "", err
	//}

	return formattedid.String(), nil

}
