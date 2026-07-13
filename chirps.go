package main

import (
	"net/http"
	"time"
	"github.com/dev-karani/chirpy/internal/auth"
	"github.com/google/uuid"
	database "github.com/dev-karani/chirpy/internal/database"
	"sort"
	"fmt"
	"encoding/json"
	"strings"
)

func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	//get token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to get token")
		return
	}

	//authenicate user
	authenticateUserID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing or invalid token")
		return
	}

	//get chirp id
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "invalid chirp id")
		return
	}

	//confirm chirp exists
	dbChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	//confirm db chripuserid == authethenitcates user
	if dbChirp.UserID != authenticateUserID {
		respondWithError(w, http.StatusForbidden, "chirp user id is not same to authenticateUserID")
		return
	}
	//delete chirp
	err = cfg.dbQueries.DeleteChirpByID(r.Context(), dbChirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "chirp delete failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (cfg *apiConfig) handlerChirpsGetByID(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, 404, "invalid chirp ID")
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	respondWithJSON(w, 200, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
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
		dbChirps, err = cfg.dbQueries.GetAllChirps(r.Context())
	} else {
		//turn query into valid uuid
		authorID, err := uuid.Parse(queryAuthorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid uuid")
		}

		//get chirp with author id
		dbChirps, _ = cfg.dbQueries.GetChirpsByAuthorID(r.Context(), authorID)
	}

	//handle error when getting dbchirps
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt retrieve chirp")
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
	respondWithJSON(w, 200, chirps)
}


type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}


type createChirpRequest struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := createChirpRequest{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't decode parameters")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error getting authbearer token")
		return
	}

	authenticatedUserID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing or invalid token")
		return
	}

	if len(req.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	cleaned := cleanBody(req.Body)

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: authenticatedUserID,
	})
	if err != nil {

		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

// --------- chirp validation ---------

type chirpRequest struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}


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