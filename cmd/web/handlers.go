package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *application) VirtualTerminalHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "terminal", nil, "stripe-js")

	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) PaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := map[string]any{
		"cardholderName":  r.Form.Get("cardholder-name"),
		"cardholderEmail": r.Form.Get("cardholder-email"),
		"amount":          r.Form.Get("payment-intent-amount"),
		"currency":        r.Form.Get("payment-intent-currency"),
		"paymentMethod":   r.Form.Get("payment-intent-method"),
		"paymentIntent":   r.Form.Get("payment-intent"),
	}

	err = app.renderTemplate(w, r, "succeeded", &templateData{Data: data})

	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) BuyWidgetHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)

	if err != nil {
		app.errorLog.Println("invalid widget id", err)
		return
	}

	widget, err := app.DBModel.GetWidget(id)

	if err != nil {
		app.errorLog.Println("failed to get widget", err)
		return
	}

	data := map[string]any{
		"widget": widget,
	}

	td := templateData{
		Data: data,
	}

	err = app.renderTemplate(w, r, "buy", &td, "stripe-js")

	if err != nil {
		app.errorLog.Println(err)
		return
	}
}
