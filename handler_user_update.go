package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/auth"
	"github.com/marioxcolomar/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Print("error when getting bearer token: \n", err)
		respondWithError(w, http.StatusUnauthorized, "request is missing JWT", err)
		return
	}
	userId, errJWT := auth.ValidateJWT(token, cfg.jwtSecret)
	if errJWT != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't validate JWT", errJWT)
		return
	}

	type RequestParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	dec := json.NewDecoder(r.Body)
	params := RequestParams{}
	errDecode := dec.Decode(&params)
	if errDecode != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", err)
		return
	}

	// Hash password
	password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to hash password", err)
		return
	}
	errUpdate := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{ID: userId, Email: params.Email, HashedPassword: password})
	if errUpdate != nil {
		respondWithError(w, http.StatusUnauthorized, "error updating user", err)
		return
	}
	type User struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}
	type Response struct {
		User
	}
	// Get user after email was updated
	user, err := cfg.dbQueries.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to find the user", nil)
		return
	}
	respondWithJSON(w, http.StatusOK, Response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
