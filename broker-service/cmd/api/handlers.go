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
	Log    LogPayload  `json:"log,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type LogPayload struct {
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
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
	case "log":
		app.logItem(w, reqPayload.Log)
	default:
		app.errorJSON(w, errors.New("Unknown action"))
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	log.Printf("Attempting to hand log to logger service\n")
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		app.errorJSON(w, err)
	}

	logServiceUrl := "http://logger-service/log"
	req, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request\n")
		app.errorJSON(w, err)
	}

	// req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Response error\n")
		app.errorJSON(w, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		log.Printf("Response not accepted\n")
		app.errorJSON(w, err)
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged data"

	app.writeJSON(w, http.StatusAccepted, payload)
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
		log.Printf("status %d\n", response.StatusCode)
		app.errorJSON(w, errors.New("Error calling auth service"))
		return
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
		log.Printf("error: %v\n", err)
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}
