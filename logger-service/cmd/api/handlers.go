package main

import (
	"log"
	"log-service/cmd/api/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {

	log.Printf("Hitting write log\n")
	// read json
	var reqPayload JSONPayload
	err := app.readJSON(w, r, &reqPayload)
	if err != nil {
		log.Printf("Json read error\n")
		return
	}

	// insert data
	event := data.LogEntry{
		Name: reqPayload.Name,
		Data: reqPayload.Data,
	}

	err = app.Models.LogEntry.Insert(event)
	if err != nil {
		log.Printf("Insert error\n")
		app.errorJSON(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}
