package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request is missing JWT", err)
		return
	}
	userId, errJWT := auth.ValidateJWT(token, cfg.jwtSecret)
	if errJWT != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't validate JWT", errJWT)
		return
	}
	// Validate user is deleting their own chirp
	chirpIdStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "missing the necessary information to complete request", nil)
		return
	}
	chirp, err := cfg.dbQueries.GetChrip(r.Context(), chirpId)
	if err != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}
	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "not allow to delete chirps that do not belong to you", nil)
		return
	}
	// Delete chirp
	errDelete := cfg.dbQueries.DeleteChirp(r.Context(), chirpId)
	if errDelete != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to delete chirp", errDelete)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
