package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/auth"
	"github.com/marioxcolomar/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
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

	type PostChirp struct {
		ID     uuid.UUID `json:"id"`
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	dec := json.NewDecoder(r.Body)
	params := PostChirp{}
	errParams := dec.Decode(&params)
	if errParams != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", errParams)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "chirp is too long", nil)
		return
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: replaceWords(params.Body), UserID: userId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, PostChirp{
		ID:     chirp.ID,
		Body:   chirp.Body,
		UserId: chirp.UserID,
	})
}

func replaceWords(s string) string {
	ws := strings.Split(s, " ")
	clean := []string{}
	match := ""
	for _, word := range ws {
		switch strings.ToLower(word) {
		case "kerfuffle", "sharbert", "fornax":
			match = "****"
		default:
			match = word
		}
		clean = append(clean, match)
	}
	res := strings.Join(clean, " ")
	return res
}
