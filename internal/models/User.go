package models

import (
	"errors"
	"strings"
)

type User struct {
	Name string
	UUID string
}

func InitilaizeUser(Username string) (*User, error) {
	if strings.TrimSpace(Username) == "" {
		return nil, errors.New("empty Username")
	}
	user := new(User)
	user.Name = Username
	user.UUID = genUuid()

	return user, nil
}
