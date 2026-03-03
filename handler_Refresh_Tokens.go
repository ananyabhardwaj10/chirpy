package main
import (
	"net/http"
	"time"

	"github.com/ananyabhardwaj10/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshTokens(w http.ResponseWriter, req *http.Request) {
	reftoken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error extracting token")
		return 
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), reftoken) 
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return 
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwt_secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error generating new access token")
		return 
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	refToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error extracting refresh token")
		return 
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), refToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error revoking refresh token")
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}

