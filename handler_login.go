package main 
import(
	"net/http"
	"encoding/json"
	"time"
	"github.com/google/uuid"
	"github.com/ananyabhardwaj10/chirpy/internal/auth"
	"github.com/ananyabhardwaj10/chirpy/internal/database"
	
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params) 
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return 
	}


	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "cannot get user by email")
		return 
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if !match || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return 
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwt_secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating JWT")
		return 
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})

	type response struct {
    	ID        uuid.UUID `json:"id"`
    	CreatedAt time.Time `json:"created_at"`
    	UpdatedAt time.Time `json:"updated_at"`
    	Email     string    `json:"email"`
		Token 	  string    `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed bool    `json:"is_chirpy_red"`
	}

	respondWithJSON(w, http.StatusOK, response{
    	ID:        user.ID,
    	CreatedAt: user.CreatedAt,
    	UpdatedAt: user.UpdatedAt,
    	Email:     user.Email,
		Token: 	   token,
		RefreshToken: refreshToken,
		IsChirpyRed: user.IsChirpyRed,
	})
}