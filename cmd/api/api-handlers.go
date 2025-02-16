package main

import (
	"encoding/json"
	"net/http"
)

type stripePayload struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Content string `json:"content"`
	ID      int    `json:"id"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	res := jsonResponse{
		OK: true,
	}

	json, err := json.MarshalIndent(res, "", "    ")

	if err != nil {
		app.errorLog.Println("error in marshalling response", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
