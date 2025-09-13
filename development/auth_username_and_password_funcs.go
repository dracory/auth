package main

import (
	"errors"

	"github.com/dracory/auth"
	"github.com/gouniverse/utils"
)

func userFindByUsername(username string, firstName string, lastName string, options auth.UserAuthOptions) (userID string, err error) {
	slug := utils.StrSlugify(username, rune('_'))
	var user map[string]string
	err = jsonStore.Read("users", slug, &user)
	if err != nil {
		return "not found err", err
	}

	if user == nil {
		return "not found", errors.New("unable to find user")
	}

	return user["id"], nil
}

func userPasswordChange(userID string, password string, options auth.UserAuthOptions) error {
	user, err := userFindByID(userID)
	if err != nil {
		return err
	}

	user["password"] = password

	slug := utils.StrSlugify(user["username"], rune('_'))
	errSave := jsonStore.Write("users", slug, user)
	if errSave != nil {
		return errSave
	}

	jsonStore.Delete("users", slug)

	return nil
}

func userRegister(username string, password string, first_name string, last_name string, options auth.UserAuthOptions) error {
	slug := utils.StrSlugify(username, rune('_'))
	err := jsonStore.Write("users", slug, map[string]string{
		"id":         utils.StrRandomFromGamma(16, "abcdef0123456789"),
		"username":   username,
		"password":   password,
		"first_name": first_name,
		"last_name":  last_name,
	})
	if err != nil {
		return err
	}
	return nil
}
