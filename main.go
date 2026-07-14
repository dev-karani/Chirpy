package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/dev-karani/chirpy/chirps"
	database "github.com/dev-karani/chirpy/internal/database"
	"github.com/dev-karani/chirpy/polka"
	users "github.com/dev-karani/chirpy/users"
)

//before refactor

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

	userHandler := users.NewHandler(apiCfg.dbQueries, apiCfg.jwtsecret)
	chirpsHandler := chirps.NewHandler(apiCfg.dbQueries, apiCfg.jwtsecret)
	polkaHandler := polka.NewHandler(apiCfg.dbQueries, apiCfg.polkaKey)

	mux := http.NewServeMux()
	registerRoutes(mux, apiCfg, userHandler, chirpsHandler, polkaHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
