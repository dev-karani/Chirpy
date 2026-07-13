package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	// "sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/dev-karani/chirpy/internal/auth"
	database "github.com/dev-karani/chirpy/internal/database"
)



//before refactor


// updateUserShape
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handle update of password
func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	updateReq := UpdateUserRequest{}
	if err := decoder.Decode(&updateReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to decode")
		return
	}

	//authenicate user
	//get bearer token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to get token")
		return
	}

	authenticatedUserID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing  or invalid token")
		return
	}
	//hashedPassword
	hashedPassword, err := auth.HashPassword(updateReq.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	//call db with values
	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          updateReq.Email,
		HashedPassword: hashedPassword,
		ID:             authenticatedUserID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

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

// createUser
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}
type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	//decode incoming json
	decoder := json.NewDecoder(r.Body)
	req := createUserRequest{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	//2.call sqlc-generated function
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("error hashing password: %s", err)
		respondWithError(w, 500, "something went wrong")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("error creating user:%s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	//respond with created user
	respondWithJSON(w, 201, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := loginRequest{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	jwtToken, err := auth.MakeJWT(user.ID, cfg.jwtsecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "failed to generate jwt")
		return
	}

	//refreshToken
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt create refresh token")
	}

	now := time.Now().UTC()

	_, err = cfg.dbQueries.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			CreatedAt: now,
			UpdatedAt: now,
			UserID:    user.ID,
			ExpiresAt: now.Add(60 * 24 * time.Hour),
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt save refresh token")
		return
	}

	respondWithJSON(w, 200, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwtToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	})
}

type RefreshResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	dbToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	jwtToken, err := auth.MakeJWT(dbToken.UserID, cfg.jwtsecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create jwt")
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: jwtToken,
	})
}

func (cfg *apiConfig) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't revoke refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func main() {
	fmt.Println("Before refactor")
	if err := godotenv.Load(); err != nil {
		// keep going if env isn't present; tests usually provide DB_URL via environment
		log.Printf("warning: could not load .env: %v", err)
	}
	//print the env variables
	fmt.Println("PLATFORM=", os.Getenv("PLATFORM"))
	fmt.Println("SECRET=", os.Getenv("SECRET"))
	fmt.Println("DB_URL=", os.Getenv("DB_URL"))
	fmt.Println("POLKA_KEY=", os.Getenv("POLKA_KEY"))

	polkaKeyEnv := os.Getenv("POLKA_KEY")
	if polkaKeyEnv == "" {
		fmt.Println("polka key missing")
	}
	jwtSecret := os.Getenv("SECRET")
	if jwtSecret == "" {
		log.Fatal("missing jwt secret")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("platform must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiCfg := &apiConfig{
		dbQueries: database.New(db),
		platform:  platform,
		jwtsecret: jwtSecret,
		polkaKey:  polkaKeyEnv,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))
	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefresh)
	//create users
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	mux.HandleFunc("POST /api/revoke", apiCfg.HandlerRevoke)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	//delete users
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirpByID)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	//a
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPostPolkaWebhook)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetByID)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)

	//login
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
