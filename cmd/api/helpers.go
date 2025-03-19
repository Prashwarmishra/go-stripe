package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	res, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	for key, value := range headers[0] {
		w.Header()[key] = value
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

	data, err := json.MarshalIndent(response, "", "\t")

	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
	return nil
}

func (app *application) invalidCredentials(w http.ResponseWriter) error {
	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	response.Error = true
	response.Message = "invalid credentials. pass correct email and password"

	data, err := json.MarshalIndent(response, "", "\t")

	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(data)

	return nil
}
