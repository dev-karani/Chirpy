package chirps

import (
	"github.com/dev-karani/chirpy/internal/auth"
	database "github.com/dev-karani/chirpy/internal/database"
	"github.com/dev-karani/chirpy/internal/httpAPI"
	"github.com/google/uuid"

	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type Handler struct {
	db        *database.Queries
	jwtSecret string
}

func NewHandler(db *database.Queries, jwtSecret string) *Handler {
	return &Handler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) DeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	//get token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "failed to get token")
		return
	}

	//authenicate user
	authenticateUserID, err := auth.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "missing or invalid token")
		return
	}

	//get chirp id
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusNotFound, "invalid chirp id")
		return
	}

	//confirm chirp exists
	dbChirp, err := h.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	//confirm db chripuserid == authethenitcates user
	if dbChirp.UserID != authenticateUserID {
		httpAPI.RespondWithError(w, http.StatusForbidden, "chirp user id is not same to authenticateUserID")
		return
	}
	//delete chirp
	err = h.db.DeleteChirpByID(r.Context(), dbChirp.ID)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "chirp delete failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) GetChirpsByID(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusNotFound, "invalid chirp ID")
		return
	}

	dbChirp, err := h.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	httpAPI.RespondWithJSON(w, 200, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (h *Handler) GetChirp(w http.ResponseWriter, r *http.Request) {
	var (
		dbChirps []database.Chirp
		err      error
	)
	//get request query
	queryAuthorID := r.URL.Query().Get("author_id")
	queryBySort := r.URL.Query().Get("sort")

	//for queryauthorID && queryBySort
	//check if empty
	if queryAuthorID == "" {
		//get all chirps
		dbChirps, err = h.db.GetAllChirps(r.Context())
	} else {
		//turn query into valid uuid
		authorID, err := uuid.Parse(queryAuthorID)
		if err != nil {
			httpAPI.RespondWithError(w, http.StatusBadRequest, "invalid uuid")
		}

		//get chirp with author id
		dbChirps, _ = h.db.GetChirpsByAuthorID(r.Context(), authorID)
	}

	//handle error when getting dbchirps
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "couldnt retrieve chirp")
		return
	}

	//turn chirps to shfitsize according to dbchirps
	chirps := make([]Chirp, 0, len(dbChirps))
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}
	fmt.Println(chirps)
	if queryBySort == "desc" {
		sort.Slice(chirps, func(i int, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})

	}
	httpAPI.RespondWithJSON(w, 200, chirps)
}

func (h *Handler) CreateChirps(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := createChirpRequest{}
	err := decoder.Decode(&req)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "Couldn't decode parameters")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "error getting authbearer token")
		return
	}

	authenticatedUserID, err := auth.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "missing or invalid token")
		return
	}

	if len(req.Body) > 140 {
		httpAPI.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")

		return
	}
	cleaned := cleanBody(req.Body)

	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: authenticatedUserID,
	})
	if err != nil {

		httpAPI.RespondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	httpAPI.RespondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

// --------- chirp validation ---------

// type rs struct {
// 	CleanedBody string `json:"cleaned_body"`
// }

func cleanBody(body string) string {
	splitWords := strings.Split(body, " ")

	badWordSlice := []string{"kerfuffle", "sharbert", "fornax"}
	for i, word := range splitWords {
		for _, badWord := range badWordSlice {
			if strings.ToLower(word) == badWord {
				splitWords[i] = "****"
			}
		}
	}
	return strings.Join(splitWords, " ")
}

