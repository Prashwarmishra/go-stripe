package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Get("/virtual-terminal", app.VirtualTerminalHandler)
	mux.Post("/payment-succeeded", app.PaymentSucceededHandler)
	mux.Get("/buy", app.BuyWidgetHandler)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
