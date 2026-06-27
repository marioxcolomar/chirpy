package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/auth"
	"github.com/marioxcolomar/chirpy/internal/database"
)

type AuthHandler struct {
	db  *database.Queries
	jwt string
}

func NewAuthHandler(db *database.Queries, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwt: jwtSecret}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	dec := json.NewDecoder(r.Body)
	params := LoginRequest{}
	err := dec.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", err)
		return
	}

	user, err := h.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, h.jwt, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to complete request", err)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to complete request", err)
		return
	}
	refresh, err := h.db.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().UTC().AddDate(0, 0, 60),
		})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to complete request", err)
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
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	respondWithJSON(w, http.StatusOK, Response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refresh.Token,
	})
}

func (h *AuthHandler) HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request missing JWT", err)
		return
	}
	refreshToken, err := h.db.GetRefreshToken(r.Context(), token)
	fmt.Println("err is not nil: \n", err != nil)
	fmt.Println("revoked at valid: \n", !refreshToken.RevokedAt.Valid)
	fmt.Println("token expires at: \n", refreshToken.ExpiresAt)
	if err != nil || refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "unable to complete request", err)
		return
	}

	newToken, err := auth.MakeJWT(refreshToken.UserID, h.jwt, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to complete request", err)
		return
	}

	type RefreshTokenResponse struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, http.StatusOK, RefreshTokenResponse{
		Token: newToken,
	})
}

func (h *AuthHandler) HandleTokenRevoke(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request missing JWT", err)
		return
	}
	errRevoke := h.db.RevokeRefreshToken(r.Context(), token)
	if errRevoke != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to handle request", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
