package main

import "net/http"

func (app *application) VirtualTerminalHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "terminal", nil)

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
