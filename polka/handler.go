package polka

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dev-karani/chirpy/internal/auth"
	database "github.com/dev-karani/chirpy/internal/database"
	"github.com/dev-karani/chirpy/internal/httpAPI"
	"github.com/google/uuid"
)

type Handler struct {
	db       *database.Queries
	polkaKey string
}

func NewHandler(db *database.Queries, polkaKey string) *Handler {
	return &Handler{
		db:       db,
		polkaKey: polkaKey,
	}
}

// post polka webhook
func (h *Handler) PostPolkaWebhook(w http.ResponseWriter, r *http.Request) {

	//check apikey in header==env one
	token, err := auth.GetAPIKey(r.Header)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "invalid token")
	}

	if token != h.polkaKey {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	polkaRes := PolkaRequest{}

	if err := decoder.Decode(&polkaRes); err != nil {
		httpAPI.RespondWithError(w, http.StatusBadRequest, "failed to decode")
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
		httpAPI.RespondWithError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	//update database
	_, err = h.db.UpgradeUserRed(r.Context(), userID)
	if err == sql.ErrNoRows {
		httpAPI.RespondWithError(w, http.StatusNotFound, "user does not exist")
		return
	}
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "failed to update column")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
