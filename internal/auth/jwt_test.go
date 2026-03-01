package auth 
import(
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestJWT1(t *testing.T) {
    secret := "testsecret"
    userID := uuid.New()

    token, err := MakeJWT(userID, secret, time.Hour)
    if err != nil {
        t.Fatalf("unexpected error creating token: %v", err)
    }

    parsedID, err := ValidateJWT(token, secret)
    if err != nil {
        t.Fatalf("unexpected error validating token: %v", err)
    }

    if parsedID != userID {
        t.Fatalf("expected %v, got %v", userID, parsedID)
    }
}

func TestJWT2(t *testing.T) {
    secret := "testing jwt working"
    userID := uuid.New()

    token, err := MakeJWT(userID, secret, time.Minute)
    if err != nil {
        t.Fatalf("unexpected error creating token: %v", err)
    }

    parsedID, err := ValidateJWT(token, secret)
    if err != nil {
        t.Fatalf("unexpected error validating token: %v", err)
    }

    if parsedID != userID {
        t.Fatalf("expected %v, got %v", userID, parsedID)
    }
}