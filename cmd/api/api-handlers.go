package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	if cust != nil {
		_, err := card.SubscribeToPlan(cust, stripePayload.Plan,
			stripePayload.Email, stripePayload.LastFour, stripePayload.CardType)

		if err != nil {
			app.errorLog.Print(err)
			transactionMessage = "Failed to subscribe to plan"
			okay = false
		}

	}

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

func (app *application) AuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, err)
		app.errorLog.Println(err)
		return
	}

	// get email from db
	user, err := app.DBModel.GetUserDetailsByEmail(payload.Email)

	if err != nil {
		app.errorLog.Println(err)
		err = app.invalidCredentials(w)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}
	app.infoLog.Println("user", user)

	// compare password
	isValid, err := app.validatePassword(user.Password, payload.Password)

	if err != nil || !isValid {
		app.errorLog.Println(err)
		err = app.invalidCredentials(w)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	// create credentials
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	// save token
	err = app.DBModel.InsertToken(user, token)

	if err != nil {
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	var jsonResponse struct {
		Okay    bool         `json:"okay"`
		Message string       `json:"message"`
		Token   models.Token `json:"authentication_token"`
		UserID  int          `json:"user_id"`
	}

	jsonResponse.Okay = true
	jsonResponse.Message = fmt.Sprintf("auth token generated for %s", user.Email)
	jsonResponse.Token = *token
	jsonResponse.UserID = user.ID

	err = app.writeJSON(w, http.StatusOK, &jsonResponse)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
	}
}

func (app *application) authenticateUser(r *http.Request) (*models.User, error) {
	authorizationHeader := r.Header.Get("Authorization")

	if authorizationHeader == "" {
		return nil, errors.New("empty authorization passed in headers")
	}

	headerParts := strings.Split(authorizationHeader, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("invalid authorization token format")
	}

	token := headerParts[1]

	if len(token) != 26 {
		return nil, errors.New("invalid authorization token length")
	}

	user, err := app.DBModel.GetUserFromToken(token)

	if err != nil {
		app.errorLog.Println(err)
		return nil, errors.New("no user mapped to this token")
	}

	return user, nil
}

func (app *application) CheckAuthentication(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticateUser(r)

	if err != nil {
		app.errorLog.Println(err)
		err = app.invalidCredentials(w)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	response.Error = false
	response.Message = fmt.Sprintf("authenticated user %s", user.Email)

	err = app.writeJSON(w, http.StatusOK, &response)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
	}
}

func (app *application) VirtualTerminalPaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	type TransactionData struct {
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		CardholderName  string `json:"cardholder_name"`
		CardholderEmail string `json:"cardholder_email"`
		PaymentAmount   int    `json:"payment_amount"`
		Currency        string `json:"currency"`
		PaymentIntent   string `json:"payment_intent"`
		PaymentMethod   string `json:"payment_method"`
		LastFour        string `json:"last_four"`
		ExpiryMonth     int    `json:"expiry_month"`
		ExpiryYear      int    `json:"expiry_year"`
		BankReturnCode  string `json:"bank_return_code"`
	}

	transactionData := TransactionData{}

	err := app.readJSON(w, r, &transactionData)

	if err != nil {
		app.errorLog.Println(err)
		err = app.badRequest(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	card := cards.Card{
		Key:    app.config.stripe.key,
		Secret: app.config.stripe.secret,
	}

	pm, err := card.GetPaymentMethod(transactionData.PaymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		err = app.badRequest(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	pi, err := card.RetrievePaymentIntent(transactionData.PaymentIntent)

	if err != nil {
		app.errorLog.Println(err)
		err = app.badRequest(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	transactionData.LastFour = pm.Card.Last4
	transactionData.ExpiryMonth = int(pm.Card.ExpMonth)
	transactionData.ExpiryYear = int(pm.Card.ExpYear)
	transactionData.BankReturnCode = pi.Charges.Data[0].ID

	customerId, err := app.SaveCustomer(transactionData.FirstName,
		transactionData.LastName, transactionData.CardholderEmail)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	transactionId, err := app.SaveTransaction(transactionData.PaymentAmount, transactionData.ExpiryMonth,
		transactionData.ExpiryYear, transactionData.Currency, transactionData.LastFour,
		transactionData.BankReturnCode, transactionData.PaymentIntent, transactionData.PaymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	_, err = app.SaveOrder(1, transactionId, customerId, transactionData.PaymentAmount)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}

	var response struct {
		Error           bool            `json:"error"`
		Message         string          `json:"message"`
		TransactionData TransactionData `json:"transaction_data"`
	}

	response.Error = false
	response.Message = "transaction successful"
	response.TransactionData = transactionData

	err = app.writeJSON(w, http.StatusOK, response)

	if err != nil {
		app.errorLog.Println(err)
		err = app.internalServerError(w, err)
		if err != nil {
			app.errorLog.Println(err)
		}
		return
	}
}
