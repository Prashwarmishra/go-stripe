package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	res, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048567
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("invalid json passed as payload")
	}

	return nil
}

func (app *application) badRequest(w http.ResponseWriter, err error) error {
	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	response.Error = true
	response.Message = err.Error()

	return app.writeJSON(w, http.StatusBadRequest, response)
}

func (app *application) invalidCredentials(w http.ResponseWriter) error {
	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	response.Error = true
	response.Message = "invalid credentials, input correct email and password"

	return app.writeJSON(w, http.StatusUnauthorized, response)
}

func (app *application) internalServerError(w http.ResponseWriter, err error) error {
	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	response.Error = true
	response.Message = fmt.Sprintln("internal server error", err)

	return app.writeJSON(w, http.StatusInternalServerError, response)
}

func (app *application) validatePassword(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
