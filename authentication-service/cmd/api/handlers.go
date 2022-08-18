package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var reqPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &reqPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		log.Printf("bad request read json\n")
		return
	}

	log.Printf("payload : \n", reqPayload)

	// validate user against db
	user, err := app.Models.User.GetByEmail(reqPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusBadRequest)
		log.Printf("invalid 1\n")
		return
	}

	valid, err := user.PasswordMatches(reqPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusBadRequest)
		log.Printf("invalid 2\n")
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in as user %s\n", user.Email),
	}
	log.Printf("reponding\n")

	app.writeJSON(w, http.StatusAccepted, payload)
}
