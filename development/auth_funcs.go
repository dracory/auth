package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/dracory/auth"
	"github.com/dracory/auth/development/scribble"
	"github.com/dracory/str"
	"github.com/spf13/cast"
)

var jsonStore *scribble.Driver

func emailSend(userID string, subject string, body string) error {
	emailSendTo("info@sinevia.com", []string{"info@sinevia.com"}, subject, body)
	return nil
}

func userLogin(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
	slug := str.Slugify(username, rune('_'))
	var user map[string]string
	err = jsonStore.Read("users", slug, &user)
	if err != nil {
		return "not found err", err
	}
	log.Println(user)
	if user == nil {
		return "not found", errors.New("unable to find user")
	}

	if user["password"] == password {
		return username, nil
	}

	return "password mismatch", errors.New("password mismatch")
}

func userLogout(username string, options auth.UserAuthOptions) error {
	return nil
}

func userFindByAuthToken(token string, options auth.UserAuthOptions) (userID string, err error) {
	slug := str.Slugify(token, rune('_'))
	err = jsonStore.Read("tokens", slug, &userID)
	if err != nil {
		return "not found err", err
	}
	return userID, nil
}

// func userFindByUsername(username string, firstName string, lastName string) (userID string, err error) {
// 	slug := utils.StrSlugify(username, rune('_'))
// 	var user map[string]string
// 	err = jsonStore.Read("users", slug, &user)
// 	if err != nil {
// 		return "not found err", err
// 	}

// 	if user == nil {
// 		return "not found", errors.New("unable to find user")
// 	}

// 	return user["id"], nil
// }

// func userPasswordChange(userID string, password string) error {
// 	user, err := userFindByID(userID)
// 	if err != nil {
// 		return err
// 	}

// 	user["password"] = password

// 	slug := utils.StrSlugify(user["username"], rune('_'))
// 	errSave := jsonStore.Write("users", slug, user)
// 	if errSave != nil {
// 		return errSave
// 	}

// 	jsonStore.Delete("users", slug)

// 	return nil
// }

func userFindByID(userID string) (user map[string]string, err error) {
	users, errReadAll := jsonStore.ReadAll("users")
	if errReadAll != nil {
		return nil, errReadAll
	}

	for _, userJson := range users {
		json.Unmarshal(userJson, &user)
		log.Println(userID)
		log.Println(user)
		if user["id"] == userID {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

func userStoreAuthToken(token string, userID string, options auth.UserAuthOptions) error {
	slug := str.Slugify(token, rune('_'))
	err := jsonStore.Write("tokens", slug, userID)
	if err != nil {
		return err
	}
	return nil
}

// func userRegister(username string, password string, first_name string, last_name string) error {
// 	slug := utils.StrSlugify(username, rune('_'))
// 	err := jsonStore.Write("users", slug, map[string]string{
// 		"id":         utils.StrRandomFromGamma(16, "abcdef0123456789"),
// 		"username":   username,
// 		"password":   password,
// 		"first_name": first_name,
// 		"last_name":  last_name,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func temporaryKeyGet(key string) (value string, err error) {
	slug := str.Slugify(key, rune('_'))
	var record map[string]string
	err = jsonStore.Read("temp", slug, &record)
	if err != nil {
		return "", err
	}
	return record["value"], nil
}

func temporaryKeySet(key string, value string, expiresSeconds int) (err error) {
	slug := str.Slugify(key, rune('_'))
	expiresAt := time.Now().Add(time.Duration(expiresSeconds))
	err = jsonStore.Write("temp", slug, map[string]string{
		"id":           str.RandomFromGamma(16, "abcdef0123456789"),
		"value":        value,
		"expires":      cast.ToString(expiresSeconds),
		"expires_time": cast.ToString(expiresAt),
	})
	if err != nil {
		return err
	}
	return nil
}
