package main

import "net/http"

func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
