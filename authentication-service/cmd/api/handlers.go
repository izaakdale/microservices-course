package main

import (
	"bytes"
	"encoding/json"
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
		app.errorJSON(w, errors.New("Invalid credentials"), http.StatusUnauthorized)
		log.Printf("invalid 2\n")
		return
	}

	// log auth req
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.errorJSON(w, err)
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in as user %s\n", user.Email),
	}
	log.Printf("reponding\n")

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name,omitempty"`
		Data string `json:"data,omitempty"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceUrl := "http://logger-service/log"

	req, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
