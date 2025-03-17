package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-stripe/internal/cards"
	"github.com/go-stripe/internal/models"
)

type TransactionData struct {
	FirstName       string
	LastName        string
	CardholderName  string
	CardholderEmail string
	PaymentAmount   int
	Currency        string
	PaymentIntent   string
	PaymentMethod   string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

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

func (app *application) GetTransactionData(r *http.Request) (*TransactionData, error) {
	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	cardholderName := r.Form.Get("cardholder-name")
	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	cardholderEmail := r.Form.Get("cardholder-email")
	amount := r.Form.Get("amount")
	currency := r.Form.Get("payment-intent-currency")
	paymentIntentId := r.Form.Get("payment-intent")
	paymentMethodId := r.Form.Get("payment-intent-method")

	card := cards.Card{
		Key:    app.config.stripe.key,
		Secret: app.config.stripe.secret,
	}

	paymentMethod, err := card.GetPaymentMethod(paymentMethodId)

	if err != nil {
		app.errorLog.Println("failed to fetch payment method", err)
		return nil, err
	}

	paymentIntent, err := card.RetrievePaymentIntent(paymentIntentId)

	if err != nil {
		app.errorLog.Println("failed to fetch payment intent", err)
		return nil, err
	}

	lastFour := paymentMethod.Card.Last4
	expiryMonth := paymentMethod.Card.ExpMonth
	expiryYear := paymentMethod.Card.ExpYear
	bankReturnCode := paymentIntent.Charges.Data[0].ID

	paymentAmount, err := strconv.Atoi(amount)

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	txnData := TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		CardholderName:  cardholderName,
		CardholderEmail: cardholderEmail,
		PaymentAmount:   paymentAmount,
		Currency:        currency,
		PaymentIntent:   paymentIntentId,
		PaymentMethod:   paymentMethodId,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  bankReturnCode,
	}

	return &txnData, nil
}

func (app *application) PaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	widgetIdStr := r.Form.Get("widget-id")

	customerId, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.CardholderEmail)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	transactionId, err := app.SaveTransaction(txnData.PaymentAmount, txnData.ExpiryMonth,
		txnData.ExpiryYear, txnData.Currency, txnData.LastFour,
		txnData.BankReturnCode, txnData.PaymentIntent, txnData.PaymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	widgetId, err := strconv.Atoi(widgetIdStr)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	_, err = app.SaveOrder(widgetId, transactionId, customerId, txnData.PaymentAmount)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

func (app *application) ReceiptHandler(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	app.Session.Remove(r.Context(), "receipt")

	data := map[string]any{
		"txn": txn,
	}

	err := app.renderTemplate(w, r, "receipt", &templateData{Data: data})

	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) VirtualTerminalPaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	_, err = app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.CardholderEmail)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	_, err = app.SaveTransaction(txnData.PaymentAmount, txnData.ExpiryMonth,
		txnData.ExpiryYear, txnData.Currency, txnData.LastFour,
		txnData.BankReturnCode, txnData.PaymentIntent, txnData.PaymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

func (app *application) VirtualTerminalReceiptHandler(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	app.Session.Remove(r.Context(), "receipt")
	data := map[string]any{
		"txn": txn,
	}

	err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{Data: data})

	if err != nil {
		app.errorLog.Println(err)
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

func (app *application) BronzePlanHandler(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DBModel.GetWidget(2)

	if err != nil {
		app.errorLog.Println(err)
	}

	data := map[string]any{
		"widget": widget,
	}

	err = app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	})

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) BronzePlanReceiptHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "bronze-plan-receipt", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "login", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}
