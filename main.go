package main 

import(
	"net/http"
	"sync/atomic"
	"fmt"
	"encoding/json"
	"log"
	"strings"
	_ "github.com/lib/pq"
	"os"
	"time"
	"database/sql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/ananyabhardwaj10/chirpy/internal/database"
	"github.com/ananyabhardwaj10/chirpy/internal/auth"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string 
}

type User struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Email      string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(
		`<html>
  			<body>
    			<h1>Welcome, Chirpy Admin</h1>
    			<p>Chirpy has been visited %d times!</p>
  			</body>
		</html>`, hits)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Only allowed in dev!"))
		return 
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteAllUsers(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusOK)

}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error":message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
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
	})
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
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

	chirp := database.Chirp{}
	chirp, err = cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body: cleaned_body,
		UserID: params.UserId,
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

func (cfg *apiConfig) handlerGetChirps (w http.ResponseWriter, req *http.Request) {
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

	type response struct {
    	ID        uuid.UUID `json:"id"`
    	CreatedAt time.Time `json:"created_at"`
    	UpdatedAt time.Time `json:"updated_at"`
    	Email     string    `json:"email"`
	}

	respondWithJSON(w, http.StatusOK, response{
    	ID:        user.ID,
    	CreatedAt: user.CreatedAt,
    	UpdatedAt: user.UpdatedAt,
    	Email:     user.Email,
	})
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("error opening database: %s", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	server := &http.Server {
		Addr: ":8080",
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := http.FileServer(http.Dir("."))

	platform := os.Getenv("PLATFORM")

	apiCfg := apiConfig {
		fileserverHits: atomic.Int32{},
		db: dbQueries,
		platform: platform,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", handler)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsers)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetSingleChirp)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	server.ListenAndServe()
}
