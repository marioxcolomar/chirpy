package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marioxcolomar/chirpy/internal/auth"
	"github.com/marioxcolomar/chirpy/internal/database"
)

type ChirpHandler struct {
	db  *database.Queries
	jwt string
}

func NewChirpHandler(db *database.Queries, jwtSecret string) *ChirpHandler {
	return &ChirpHandler{db: db, jwt: jwtSecret}
}

func (h *ChirpHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request is missing JWT", err)
		return
	}
	userId, errJWT := auth.ValidateJWT(token, h.jwt)
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
	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: replaceWords(params.Body), UserID: userId})
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

func (h *ChirpHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}
	res, err := h.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create chirp", err)
		return
	}

	out := make([]Chirp, len(res))
	for i, chirp := range res {
		out[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, out)
}

func (h *ChirpHandler) HandleGetOne(w http.ResponseWriter, r *http.Request) {
	chirpIdStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "missing the necessary information to complete request", nil)
		return
	}
	chirp, err := h.db.GetChrip(r.Context(), chirpId)
	if err != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func (h *ChirpHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	// Validate request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "request is missing JWT", err)
		return
	}
	userId, errJWT := auth.ValidateJWT(token, h.jwt)
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
	chirp, err := h.db.GetChrip(r.Context(), chirpId)
	if err != nil {
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}
	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "not allow to delete chirps that do not belong to you", nil)
		return
	}
	// Delete chirp
	errDelete := h.db.DeleteChirp(r.Context(), chirpId)
	if errDelete != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to delete chirp", errDelete)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
