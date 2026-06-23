package main

import (
	"net/http"

	"github.com/marioxcolomar/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerTokenRevoke(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request missing JWT", err)
		return
	}
	errRevoke := cfg.dbQueries.RevokeRefreshToken(r.Context(), token)
	if errRevoke != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to handle request", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
