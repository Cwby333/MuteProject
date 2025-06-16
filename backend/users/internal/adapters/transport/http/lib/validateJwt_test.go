package lib

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Тест для функции ValidateJWT
func TestValidateJWT(t *testing.T) {
	const secretKey = "my-secret-key"
	os.Setenv("JWT_SECRET_KEY", secretKey)
	defer os.Unsetenv("JWT_SECRET_KEY")

	issuer := "my-issuer"

	type testCase struct {
		description string
		tokenString string
		wantErr     bool
		wantClaims  jwt.MapClaims
	}

	testCases := []testCase{
		{
			description: "Valid token with correct signature",
			tokenString: generateValidToken(t, issuer, secretKey),
			wantErr:     false,
			wantClaims: jwt.MapClaims{
				"iss": issuer,
				"exp": float64(time.Now().Add(time.Hour).Unix()),
			},
		},
		{
			description: "Invalid token with wrong signature",
			tokenString: generateValidToken(t, issuer, "another-secret"),
			wantErr:     true,
		},
		{
			description: "Expired token",
			tokenString: generateExpiredToken(t, issuer, secretKey),
			wantErr:     true,
		},
		{
			description: "Malformed token",
			tokenString: "malformed.token.string",
			wantErr:     true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			gotClaims, err := ValidateJWT(tt.tokenString)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error state: %v, got error: %v", tt.wantErr, err)
			}

			if !tt.wantErr && !reflect.DeepEqual(gotClaims, tt.wantClaims) {
				t.Fatalf("Incorrect claims returned.\nWanted: %+v\nGot: %+v", tt.wantClaims, gotClaims)
			}
		})
	}
}

// Helper function to generate a valid JWT token
func generateValidToken(t *testing.T, issuer, key string) string {
	now := time.Now()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iss"] = issuer
	claims["exp"] = now.Add(time.Hour).Unix()
	signedToken, err := token.SignedString([]byte(key))
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}
	return signedToken
}

// Helper function to generate an expired JWT token
func generateExpiredToken(t *testing.T, issuer, key string) string {
	now := time.Now()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iss"] = issuer
	claims["exp"] = now.Add(-time.Hour).Unix()
	signedToken, err := token.SignedString([]byte(key))
	if err != nil {
		t.Fatal("Failed to sign token:", err)
	}
	return signedToken
}
