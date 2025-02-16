package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	cards "github.com/go-stripe/internal/cards"
)

type stripePayload struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	stripePayload := stripePayload{}

	err := json.NewDecoder(r.Body).Decode(&stripePayload)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(stripePayload.Amount)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Key:      app.config.stripe.key,
		Secret:   app.config.stripe.secret,
		Currency: stripePayload.Currency,
	}

	pi, errMsg, err := card.ChargeCard(card.Currency, amount)

	var res any

	if err != nil {
		res = jsonResponse{
			OK:      false,
			Message: errMsg,
		}
	} else {
		res = pi
	}

	json, err := json.MarshalIndent(res, "", "   ")

	if err != nil {
		app.errorLog.Println("error in marshalling response", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
