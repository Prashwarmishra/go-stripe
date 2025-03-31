package main

import "net/http"

func (app *application) Auth(next http.Handler) http.Handler {
	app.infoLog.Println("inside auth func")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Println("inside handler func")
		_, err := app.authenticateUser(r)
		if err != nil {
			app.errorLog.Println(err)
			err = app.invalidCredentials(w)
			if err != nil {
				app.errorLog.Println(err)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}
