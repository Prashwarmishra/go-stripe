package main

import "net/http"

func (app *application) VirtualTerminalHandler(w http.ResponseWriter, r *http.Request) {
	app.renderTemplate(w, r, "terminal", nil)
}
