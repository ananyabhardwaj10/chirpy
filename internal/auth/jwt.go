package auth 
import (
	"time"
	"fmt"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
		jwt.RegisteredClaims
	}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := MyClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Issuer: "chirpy-access",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject: userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	return token.SignedString([]byte(tokenSecret))

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := MyClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
	if token.Method != jwt.SigningMethodHS256 {
		return uuid.Nil, fmt.Errorf("Method mismatch.")
	}
	return []byte(tokenSecret), nil 
	})

	if err != nil {
		return uuid.Nil, err 
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err 
	}

	idUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return idUUID, nil 

}



