package auth
import(
	"crypto/rand"
	"encoding/hex"
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