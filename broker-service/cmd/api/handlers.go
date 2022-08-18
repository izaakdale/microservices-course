package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker!",
	}

	err := app.writeJSON(w, http.StatusOK, payload)
	if err != nil {
		log.Fatal("Failed to write json response")
	}
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload

	err := app.readJSON(w, r, &reqPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	log.Printf("payload action: %s\n", reqPayload.Action)

	switch reqPayload.Action {
	case "auth":
		app.authenticate(w, reqPayload.Auth)
	default:
		app.errorJSON(w, errors.New("Unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {

	log.Printf("Hitting authenticate\n")
	// create json to send to auth ms
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	// call service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("hitting error when creating new request\n")
		app.errorJSON(w, err)
		return
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("error client do\n")
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()
	// make sure we get correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("Invalid Credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("Error calling auth service"))
	}

	// create variable to read response body
	var jsonFromService jsonResponse
	// decode json
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		log.Printf("error decoding\n")
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Printf("error came from service\n")
		app.errorJSON(w, err, http.StatusUnauthorized)
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}
