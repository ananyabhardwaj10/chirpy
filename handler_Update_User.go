package main 
import(
	"encoding/json"
	"net/http"
	"time"
	"github.com/google/uuid"
	"github.com/ananyabhardwaj10/chirpy/internal/auth"
	"github.com/ananyabhardwaj10/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error decoding the request")
		return 
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return 
	}

	user_id, err := auth.ValidateJWT(token, cfg.jwt_secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access")
		return 
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error hashing the password")
		return 
	}

	updatedUser, err := cfg.db.UpdateUser(req.Context(), database.UpdateUserParams{
		Email: params.Email,
		HashedPassword: hashed_password,
		ID: user_id,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating user email and password")
		return 
	}

	type response struct {
    	ID        uuid.UUID `json:"id"`
    	CreatedAt time.Time `json:"created_at"`
    	UpdatedAt time.Time `json:"updated_at"`
    	Email     string    `json:"email"`
	}

	respondWithJSON(w, http.StatusOK, response{
		ID: user_id,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email: params.Email,
	})

}