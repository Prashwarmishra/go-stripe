package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-stripe/internal/cards"
	"github.com/go-stripe/internal/models"
)

func (app *application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "home", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) VirtualTerminalHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "terminal", nil, "stripe-js")

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) PaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	cardholderName := r.Form.Get("cardholder-name")
	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	cardholderEmail := r.Form.Get("cardholder-email")
	amount := r.Form.Get("amount")
	currency := r.Form.Get("payment-intent-currency")
	paymentIntentId := r.Form.Get("payment-intent")
	paymentMethodId := r.Form.Get("payment-intent-method")
	widgetIdStr := r.Form.Get("widget-id")

	data := map[string]any{
		"cardholderName":  cardholderName,
		"cardholderEmail": cardholderEmail,
		"amount":          amount,
		"currency":        currency,
		"paymentMethod":   paymentMethodId,
		"paymentIntent":   paymentIntentId,
	}

	card := cards.Card{
		Key:    app.config.stripe.key,
		Secret: app.config.stripe.secret,
	}

	paymentMethod, err := card.GetPaymentMethod(paymentMethodId)

	if err != nil {
		app.errorLog.Println("failed to fetch payment method", err)
		return
	}

	paymentIntent, err := card.RetrievePaymentIntent(paymentIntentId)

	if err != nil {
		app.errorLog.Println("failed to fetch payment intent", err)
		return
	}

	lastFour := paymentMethod.Card.Last4
	expiryMonth := paymentMethod.Card.ExpMonth
	expiryYear := paymentMethod.Card.ExpYear
	bankReturnCode := paymentIntent.Charges.Data[0].ID

	data["lastFour"] = lastFour
	data["expiryMonth"] = expiryMonth
	data["expiryYear"] = expiryYear
	data["bankReturnCode"] = bankReturnCode

	customerId, err := app.SaveCustomer(firstName, lastName, cardholderEmail)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	paymentAmount, err := strconv.Atoi(amount)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	transactionId, err := app.SaveTransaction(paymentAmount, int(expiryMonth), int(expiryYear), paymentIntentId, paymentMethodId, currency, lastFour, bankReturnCode)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	widgetId, err := strconv.Atoi(widgetIdStr)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	orderId, err := app.SaveOrder(widgetId, transactionId, customerId, paymentAmount)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println("orderId", orderId)

	app.Session.Put(r.Context(), "receipt", data)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

func (app *application) ReceiptHandler(w http.ResponseWriter, r *http.Request) {
	data := app.Session.Get(r.Context(), "receipt").(map[string]any)
	app.Session.Remove(r.Context(), "receipt")

	err := app.renderTemplate(w, r, "receipt", &templateData{Data: data})

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
