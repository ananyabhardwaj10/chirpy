package main
import(
	"encoding/json"
	"time"
	"net/http"
	"github.com/google/uuid"
	"github.com/ananyabhardwaj10/chirpy/internal/database"
	"github.com/ananyabhardwaj10/chirpy/internal/auth"
)

type User struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	IsChirpyRed bool     `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	params := parameters{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error getting user email")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error with password")
		return 
	}
	user := database.User{}
	user, err = cfg.db.CreateUser(req.Context(), database.CreateUserParams{
    	Email:          params.Email,
    	HashedPassword: hashedPassword,
	})
	respondWithJSON(w, 201, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}