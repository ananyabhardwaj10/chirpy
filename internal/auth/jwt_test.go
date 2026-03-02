package auth 
import(
	"testing"
	"time"
    "net/http"
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

func TestGetBearerToken(t *testing.T) {
	header1 := http.Header{}
	header2 := http.Header{}

	header1.Set("Authorization", "Bearer this is my tokenString")

	token1, err1 := GetBearerToken(header1)
	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}

	if token1 != "this is my tokenString" {
		t.Errorf("expected 'this is my tokenString', got '%s'", token1)
	}

	_, err2 := GetBearerToken(header2)
	if err2 == nil {
		t.Errorf("expected error for missing Authorization header")
	}
}