package users

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dev-karani/chirpy/internal/auth"
	database "github.com/dev-karani/chirpy/internal/database"
	"github.com/dev-karani/chirpy/internal/httpAPI"
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

// Create User Handler
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	//decode incoming json
	decoder := json.NewDecoder(r.Body)
	req := createUserRequest{}
	if err := decoder.Decode(&req); err != nil {
		httpAPI.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	//2.call sqlc-generated function
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("error hashing password: %s", err)
		httpAPI.RespondWithError(w, 500, "something went wrong")
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("error creating user:%s", err)
		httpAPI.RespondWithError(w, 500, "Something went wrong")
		return
	}

	//respond with created user
	httpAPI.RespondWithJSON(w, 201, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

// Login Handler
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := loginRequest{}
	if err := decoder.Decode(&req); err != nil {
		httpAPI.RespondWithError(w, http.StatusBadRequest, "Invalid reuest body")
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil || !match {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	jwtToken, err := auth.MakeJWT(user.ID, h.jwtSecret, time.Hour)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "failed to generate jwt")
		return
	}

	//refreshToken
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "couldnt create refresh token")
	}

	now := time.Now().UTC()

	_, err = h.db.CreateRefreshToken(
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
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "couldnt save refresh token")
		return
	}

	httpAPI.RespondWithJSON(w, 200, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwtToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	})
}

//

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	dbToken, err := h.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	jwtToken, err := auth.MakeJWT(dbToken.UserID, h.jwtSecret, time.Hour)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "couldn't create jwt")
		return
	}

	httpAPI.RespondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: jwtToken,
	})
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	err = h.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "couldn't revoke refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handle update of password
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	updateReq := updateUserRequest{}
	if err := decoder.Decode(&updateReq); err != nil {
		httpAPI.RespondWithError(w, http.StatusBadRequest, "failed to decode")
		return
	}

	//authenicate user
	//get bearer token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "failed to get token")
		return
	}

	authenticatedUserID, err := auth.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusUnauthorized, "missing  or invalid token")
		return
	}
	//hashedPassword
	hashedPassword, err := auth.HashPassword(updateReq.Password)
	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	//call db with values
	user, err := h.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          updateReq.Email,
		HashedPassword: hashedPassword,
		ID:             authenticatedUserID,
	})

	if err != nil {
		httpAPI.RespondWithError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	httpAPI.RespondWithJSON(w, http.StatusOK, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

