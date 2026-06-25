package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	database "github.com/dev-karani/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	html := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(html))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform !="dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err !=nil {
		log.Printf("error deleting users %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Hits reset to 0, all users delted"))
}


//createUser
type User struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Email string `json:"email"`
}
type createUserRequest struct {
	Email string `json:"email"`
}
func (cfg *apiConfig)handlerCreateUser( w http.ResponseWriter, r *http.Request){
	//decode incoming json
	decoder := json.NewDecoder(r.Body)
	req := createUserRequest{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}


	//2.call sqlc-generated function
	user, err := cfg.dbQueries.CreateUser(r.Context(),req.Email)
	if err != nil {
		log.Printf("error creating user:%s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	//respond with created user
	respondWithJSON(w, 201, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request){
	decoder := json.NewDecoder(r.Body)
	reqBody := chirpRequest{}

	err := decoder.Decode(&reqBody)
	if err != nil {
		log.Printf("error decoding request body: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleaned := cleanBody(reqBody.Body)
	
	respondWithJSON(w, 200, rs{CleanedBody: cleaned})
}



// --------- chirp validation ---------

type chirpRequest struct {
	Body string `json:"body"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type rs struct {
	CleanedBody string `json:"cleaned_body"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, errorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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
	if err := godotenv.Load(); err != nil {
		// keep going if env isn't present; tests usually provide DB_URL via environment
		log.Printf("warning: could not load .env: %v", err)
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
		platform: platform,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))

	//create users
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	//
	mux.HandleFunc("POST /admin/metrics", apiCfg.handlerMetrics)

	//delete users
	mux.HandleFunc("DELETE /admin/reset", apiCfg.handlerReset)

	// mux.HandleFunc("POST /api/chirps",)

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
