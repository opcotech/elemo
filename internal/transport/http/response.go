package http

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrMarshal = errors.New("failed to marshal response") // failed to marshal response
)

func mustWrite(w http.ResponseWriter, data []byte) {
	if _, err := w.Write(data); err != nil {
		panic(err)
	}
}

// WriteJSONResponse writes the JSON response to the response writer.
func WriteJSONResponse(w http.ResponseWriter, response any, status int) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		mustWrite(w, []byte(errors.Join(ErrMarshal, err).Error()))
	}

	w.WriteHeader(status)
	mustWrite(w, resp)
}
