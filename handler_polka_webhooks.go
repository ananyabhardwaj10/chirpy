package main
import(
	"net/http"
	"encoding/json"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ananyabhardwaj10/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerPolkaWebHooks(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error decoding request")
		return 
	}

	api_key, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error extracting the api key")
		return 
	}

	if api_key != cfg.polka_key {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return 
	}
	
	_, err = cfg.db.UserChirpyRed(req.Context(), params.Data.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Not Found")
		return 
		} else {
			respondWithError(w, http.StatusInternalServerError, "An error occurred")
			return 
		}	
	}

	w.WriteHeader(http.StatusNoContent)

}