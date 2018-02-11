package main

import (
	"github.com/satori/go.uuid"
)

// GetUUID function
func GetUUID() (id string, err error) {

	formattedid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return formattedid.String(), nil

}
