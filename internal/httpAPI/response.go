package httpAPI

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {

	RespondWithJSON(w, code, errorResponse{Error: msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling json %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

