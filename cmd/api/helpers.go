package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

type envelope map[string]any

// Sends a JSON response with a HTTP status code, some data and a header map containing
// any additional HTTP headers we want to include.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	j, err := json.Marshal(data) // json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	j = append(j, '\n') // to make it easier to view in terminal apps

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(j)

	return nil
}
