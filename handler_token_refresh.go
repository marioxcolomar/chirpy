package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/marioxcolomar/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request missing JWT", err)
		return
	}
	refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), token)
	fmt.Println("err is not nil: \n", err != nil)
	fmt.Println("revoked at valid: \n", !refreshToken.RevokedAt.Valid)
	fmt.Println("token expires at: \n", refreshToken.ExpiresAt)
	if err != nil || refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "unable to complete request", err)
		return
	}

	newToken, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret, time.Hour)
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
