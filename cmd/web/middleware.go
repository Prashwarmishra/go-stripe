package main

import "net/http"

func LoadSession(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
