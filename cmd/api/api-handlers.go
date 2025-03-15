package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	cards "github.com/go-stripe/internal/cards"
	"github.com/go-stripe/internal/models"
)

type stripePayload struct {
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Plan          string `json:"plan"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	LastFour      string `json:"last_four"`
	ExpiryMonth   int    `json:"expiry_month"`
	ExpiryYear    int    `json:"expiry_year"`
	CardType      string `json:"card_type"`
	WidgetId      string `json:"widget_id"`
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

func (app *application) GetWidgetHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)

	if err != nil {
		app.errorLog.Println("invalid widget id", err)
		return
	}

	widget, err := app.DBModel.GetWidget(id)

	if err != nil {
		app.errorLog.Println("failed to get widget from database", err)
		return
	}

	json, err := json.MarshalIndent(widget, "", "  ")

	if err != nil {
		app.errorLog.Println("failed to marshal widget json", err)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.Write(json)
}

func (app *application) CreateCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var stripePayload stripePayload
	err := json.NewDecoder(r.Body).Decode(&stripePayload)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println("stripePayload", stripePayload)

	card := cards.Card{
		Key:    app.config.stripe.key,
		Secret: app.config.stripe.secret,
	}

	cust, message, err := card.CreateCustomer(stripePayload.PaymentMethod,
		stripePayload.Email)

	okay := true
	transactionMessage := "Payment processed successfully!"

	if err != nil {
		app.errorLog.Println(message, err)
		transactionMessage = message
		okay = false
	}

	subscriptionID, err := card.SubscribeToPlan(cust, stripePayload.Plan,
		stripePayload.Email, stripePayload.LastFour, "")

	if err != nil {
		app.errorLog.Print(err)
		transactionMessage = "Failed to subscribe to plan"
		okay = false
	}

	app.infoLog.Println("subscriptionID", subscriptionID)

	res := jsonResponse{
		OK:      okay,
		Message: transactionMessage,
	}

	if okay {
		customerId, err := app.SaveCustomer(stripePayload.FirstName,
			stripePayload.LastName, stripePayload.Email)

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		amount, err := strconv.Atoi(stripePayload.Amount)

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		transactionId, err := app.SaveTransaction(amount, stripePayload.ExpiryMonth,
			stripePayload.ExpiryYear, stripePayload.Currency, stripePayload.LastFour,
			"", "", stripePayload.PaymentMethod)

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		widgetId, err := strconv.Atoi(stripePayload.WidgetId)

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		_, err = app.SaveOrder(widgetId, transactionId, customerId, amount)

		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	data, err := json.MarshalIndent(res, "", "   ")

	if err != nil {
		app.errorLog.Print(err)
		return
	}

	w.Header().Set("Accept", "application/json")
	w.Write(data)
}

func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	cx := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	customerId, err := app.DBModel.InsertCustomer(cx)

	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return customerId, err
}

func (app *application) SaveTransaction(amount, expiryMonth, expiryYear int, currency, lastFour, bankReturnCode, paymentIntent, paymentMethod string) (int, error) {
	txn := models.Transaction{
		Amount:              amount,
		Currency:            currency,
		LastFour:            lastFour,
		BankReturnCode:      bankReturnCode,
		TransactionStatusID: 2,
		ExpiryMonth:         expiryMonth,
		ExpiryYear:          expiryYear,
		PaymentIntent:       paymentIntent,
		PaymentMethod:       paymentMethod,
	}

	transactionId, err := app.DBModel.InsertTransaction(txn)

	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return transactionId, err
}

func (app *application) SaveOrder(widgetId, transactionId, customerId, amount int) (int, error) {
	order := models.Order{
		WidgetID:      widgetId,
		TransactionID: transactionId,
		StatusID:      1,
		CustomerID:    customerId,
		Quantity:      1,
		Amount:        amount,
	}

	orderId, err := app.DBModel.InsertOrder(order)

	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return orderId, err
}
