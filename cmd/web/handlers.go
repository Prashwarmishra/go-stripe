package main

import (
	"net/http"

	"github.com/go-stripe/internal/models"
)

func (app *application) VirtualTerminalHandler(w http.ResponseWriter, r *http.Request) {
	stringMap := map[string]string{
		"stripe_key": app.config.stripe.key,
	}

	err := app.renderTemplate(w, r, "terminal", &templateData{
		API:       app.config.api,
		StringMap: stringMap,
	}, "stripe-js")

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
	widget := models.Widget{
		ID:             1,
		Name:           "Custom Widget",
		Description:    "A very nice widget",
		Price:          1000,
		InventoryLevel: 10,
	}

	data := map[string]any{
		"widget": widget,
	}

	td := templateData{
		Data: data,
	}

	err := app.renderTemplate(w, r, "buy", &td, "stripe-js")

	if err != nil {
		app.errorLog.Println(err)
		return
	}
}
