package main

import (
	"sync/atomic"
	database "github.com/dev-karani/chirpy/internal/database"

)
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtsecret      string
	polkaKey       string
}