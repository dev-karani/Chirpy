package main

import (
	"net/http"
	"github.com/dev-karani/chirpy/internal/auth"
	"encoding/json"
	"github.com/google/uuid"
	"database/sql"


)

type PolkaRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

// post polka webhook
func (cfg *apiConfig) handlerPostPolkaWebhook(w http.ResponseWriter, r *http.Request) {

	//check apikey in header==env one
	token, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
	}

	if token != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	polkaRes := PolkaRequest{}

	if err := decoder.Decode(&polkaRes); err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to decode")
		return
	}

	//check type of event
	if polkaRes.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//parse uuid
	userID, err := uuid.Parse(polkaRes.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	//update database
	_, err = cfg.dbQueries.UpgradeUserRed(r.Context(), userID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "user does not exist")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update column")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
