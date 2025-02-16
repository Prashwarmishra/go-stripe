package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
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
