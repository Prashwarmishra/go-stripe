package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

type Card struct {
	Key      string
	Secret   string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) ChargeCard(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.createPaymentIntent(currency, amount)
}

func (c *Card) createPaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	params.AddMetadata("payment-source", "web")

	pi, err := paymentintent.New(params)

	if err != nil {
		msg := ""
		stripeError, ok := err.(*stripe.Error)
		if ok {
			msg = getErrorMessageFromStripeStatusCode(stripeError.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

func (c *Card) GetPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	return paymentmethod.Get(id, nil)
}

func (c *Card) RetrievePaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	return paymentintent.Get(id, nil)
}

func (c *Card) CreateCustomer(paymentMethod, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret

	params := &stripe.CustomerParams{
		PaymentMethod: stripe.String(paymentMethod),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethod),
		},
	}

	cust, err := customer.New(params)

	if err != nil {
		msg := ""
		stripeError, ok := err.(*stripe.Error)

		if ok {
			msg = getErrorMessageFromStripeStatusCode(stripeError.Code)
		}

		return nil, msg, err
	}

	return cust, "", nil
}

func (c *Card) SubscribeToPlan(cust *stripe.Customer, plan, email, lastFour, cardType string) (string, error) {
	items := []*stripe.SubscriptionItemsParams{
		{
			Plan: stripe.String(plan),
		},
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(cust.ID),
		Items:    items,
	}

	params.AddMetadata("last_four", lastFour)
	params.AddMetadata("card_type", cardType)
	params.AddExpand("latest_invoice.payment_intent")

	subscription, err := sub.New(params)

	if err != nil {
		return "", err
	}

	return subscription.ID, nil
}

func getErrorMessageFromStripeStatusCode(code stripe.ErrorCode) string {
	switch code {
	case stripe.ErrorCodeCardDeclined:
		return "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		return "Your card has expired"
	default:
		return "Your bank's declined your payment request"
	}
}
