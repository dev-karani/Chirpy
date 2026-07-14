package users

import (
	database "github.com/dev-karani/chirpy/internal/database"

)

type Handler struct {
	db *database.Queries
	jwtSecret string
}

func NewHandler(db *database.Queries,jwtSecret string) *Handler{
	return &andler{
		db:db,
		jwtSecret:jwtSecret,
	}
}