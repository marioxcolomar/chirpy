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

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	type CreateRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type User struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}
	type CreatedUserResponse struct {
		User
	}
	dec := json.NewDecoder(r.Body)
	params := CreateRequest{}
	err := dec.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", err)
		return
	}

	// Hash password
	password, err := auth.HashPassword(params.Password)
	if err != nil {
		fmt.Println("Unable to hash password: \n", err)
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: password})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, CreatedUserResponse{
		User: User{
			ID:          user.ID,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
