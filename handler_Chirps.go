package main 
import(
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ananyabhardwaj10/chirpy/internal/database"
	"github.com/ananyabhardwaj10/chirpy/internal/auth"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Something went wrong")
		return 
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return 
	}
	cleaned_body := replaceProfane(w, params.Body)

	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token string")
		return 
	}

	id, err := auth.ValidateJWT(tokenString, cfg.jwt_secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirp := database.Chirp{}
	chirp, err = cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body: cleaned_body,
		UserID: id,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating chirp")
		return 
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}

func replaceProfane(w http.ResponseWriter, message string) string {
	words := strings.Split(message, " ")
	for idx, word := range words {
		word = strings.ToLower(word)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[idx] = "****"
		}
	}

	cleaned_message := strings.Join(words, " ")
	return cleaned_message
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetAllChirps(req.Context()) 
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error retrieving all chirps")
		return 
	}

	var resp []Chirp
	for _, chirp := range chirps {
		resp = append(resp, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, req *http.Request) {
	chirp_id_str := req.PathValue("chirpID")
	chirp_id, err := uuid.Parse(chirp_id_str)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error with chirp id")
		return 
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirp_id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "no chirp by the given id")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
	
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error extracting access token")
		return 
	}

	user_id, err := auth.ValidateJWT(token, cfg.jwt_secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not validate access token")
		return 
	}

	chirp_id_str := req.PathValue("chirpID")
	chirp_id, err := uuid.Parse(chirp_id_str)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error with chirp id")
		return 
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirp_id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return 
	}

	if user_id != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Access to delete forbidden")
		return 
	}

	err = cfg.db.DeleteChirp(req.Context(), chirp_id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error deleting the chirp")
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}