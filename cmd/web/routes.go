package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(LoadSession)

	mux.Get("/", app.HomeHandler)

	mux.Get("/widget/{id}", app.BuyWidgetHandler)
	mux.Post("/payment-succeeded", app.PaymentSucceededHandler)
	mux.Get("/receipt", app.ReceiptHandler)

	mux.Get("/virtual-terminal", app.VirtualTerminalHandler)
	mux.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceededHandler)
	mux.Get("/virtual-terminal-receipt", app.VirtualTerminalReceiptHandler)

	mux.Get("/plans/bronze", app.BronzePlanHandler)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
