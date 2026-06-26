package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, errToken := auth.GetAPIKey(r.Header)
	if errToken != nil {
		respondWithError(w, http.StatusUnauthorized, "request is missing JWT", errToken)
		return
	}
	if token != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, "request is using the wrong token", nil)
		return
	}
	type Request struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	dec := json.NewDecoder(r.Body)
	params := Request{}
	errParams := dec.Decode(&params)
	if errParams != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", errParams)
		return
	}
	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}
	// Update user
	errUpdate := cfg.dbQueries.UpdateUserIsChirpyRed(r.Context(), params.Data.UserId)
	if errUpdate != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update user", errUpdate)
		return
	}
	// Find user
	_, err := cfg.dbQueries.GetUserById(r.Context(), params.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "unable to find user", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
