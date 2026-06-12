package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type PostChirp struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	dec := json.NewDecoder(r.Body)
	params := PostChirp{}
	err := dec.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "chirp is too long", nil)
		return
	}
	clean := replaceWords(params.Body)
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: clean, UserID: params.UserId})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, PostChirp{
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
