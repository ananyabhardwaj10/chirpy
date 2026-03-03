package auth
import(
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"fmt"
	"strings"
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err 
	}

	return hashed_password, nil 
}

func CheckPasswordHash(password, hash string) (bool, error) {
	password_match, err := argon2id.ComparePasswordAndHash(password, hash)
	return password_match, err 
}

func MakeRefreshToken() string {
	encoded_str := make([]byte, 32)
	rand.Read(encoded_str)

	str := hex.EncodeToString(encoded_str)
	return str
}

func GetAPIKey(headers http.Header) (string, error) {
	key := headers.Get("Authorization")

	if key == "" {
		return "", fmt.Errorf("No auth information found")
	}

	key_trimmed_prefix := strings.TrimPrefix(key, "ApiKey")

	apiKey := strings.TrimSpace(key_trimmed_prefix)

	return apiKey, nil
}