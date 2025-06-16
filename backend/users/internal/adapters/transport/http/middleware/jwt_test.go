package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateValidToken(t *testing.T, issuer, key string, tokenType string) string {
	now := time.Now()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iss"] = issuer
	claims["exp"] = now.Add(time.Hour).Unix()
	if tokenType != "" {
		claims["type"] = tokenType
	}
	signedToken, err := token.SignedString([]byte(key))
	require.NoError(t, err)
	return signedToken
}

func TestAccessJWT(t *testing.T) {
	const secretKey = "my-secret-key"
	os.Setenv("JWT_SECRET_KEY", secretKey)
	defer os.Unsetenv("JWT_SECRET_KEY")

	issuer := "my-issuer"
	validToken := generateValidToken(t, issuer, secretKey, "access")
	validRefreshToken := generateValidToken(t, issuer, secretKey, "refresh")
	badToken := "this-is-an-invalid-token"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value("claims").(jwt.MapClaims)
		if !ok {
			http.Error(w, "no claims in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AccessJWT(nextHandler)

	testCases := []struct {
		name                  string
		header                http.Header
		expectedStatus        int
		expectedTokenType     string
		expectedResponse      string
		expectClaimsInContext bool
	}{
		{
			name: "Valid token provided",
			header: http.Header{
				"Authorization": {"Bearer " + validToken},
			},
			expectedStatus:        http.StatusOK,
			expectClaimsInContext: true,
		},
		{
			name:                  "No token provided",
			header:                http.Header{},
			expectedStatus:        http.StatusUnauthorized,
			expectedResponse:      "{\"message\":\"missing auth token\",\"status_code\":401}\n",
			expectClaimsInContext: false,
		},
		{
			name: "Invalid token provided",
			header: http.Header{
				"Authorization": {"Bearer " + badToken},
			},
			expectedStatus:        http.StatusUnauthorized,
			expectedResponse:      "{\"message\":\"unauthorized\",\"status_code\":401}\n",
			expectClaimsInContext: false,
		},
		{
			name: "Refresh token instead of access token",
			header: http.Header{
				"Authorization": {"Bearer " + validRefreshToken},
			},
			expectedStatus:        http.StatusInternalServerError,
			expectedResponse:      "{\"message\":\"wrong token type\",\"status_code\":500}\n",
			expectClaimsInContext: false,
		},
		{
			name: "Missing token type",
			header: http.Header{
				"Authorization": {"Bearer " + generateValidToken(t, issuer, secretKey, "")},
			},
			expectedStatus:        http.StatusInternalServerError,
			expectedResponse:      "{\"message\":\"server error\",\"status_code\":500}\n",
			expectClaimsInContext: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header = tc.header

			recorder := httptest.NewRecorder()

			middleware.ServeHTTP(recorder, req)

			assert.Equal(t, recorder.Code, tc.expectedStatus)

			actualResponse := recorder.Body.String()
			if actualResponse != "" {
				assert.JSONEq(t, actualResponse, tc.expectedResponse)
			}

		})
	}
}

func TestRefreshJWT(t *testing.T) {
	const secretKey = "my-secret-key"
	os.Setenv("JWT_SECRET_KEY", secretKey)
	defer os.Unsetenv("JWT_SECRET_KEY")

	issuer := "my-issuer"
	validRefreshToken := generateValidToken(t, issuer, secretKey, "refresh")
	validAccessToken := generateValidToken(t, issuer, secretKey, "access")
	badToken := "this-is-an-invalid-token"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value("claims").(jwt.MapClaims)
		if !ok {
			http.Error(w, "no claims in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RefreshJWT(nextHandler)

	testCases := []struct {
		name                  string
		header                http.Header
		expectedStatus        int
		expectedResponse      string
		expectClaimsInContext bool
	}{
		{
			name: "Valid refresh token provided",
			header: http.Header{
				"Authorization": {"Bearer " + validRefreshToken},
			},
			expectedStatus:        http.StatusOK,
			expectedResponse:      "",
			expectClaimsInContext: true,
		},
		{
			name:                  "No token provided",
			header:                http.Header{},
			expectedStatus:        http.StatusUnauthorized,
			expectedResponse:      "{\"message\":\"missing auth token\",\"status_code\":401}\n",
			expectClaimsInContext: false,
		},
		{
			name: "Invalid token provided",
			header: http.Header{
				"Authorization": {"Bearer " + badToken},
			},
			expectedStatus:        http.StatusUnauthorized,
			expectedResponse:      "{\"message\":\"unauthorized\",\"status_code\":401}\n",
			expectClaimsInContext: false,
		},
		{
			name: "Access token instead of refresh token",
			header: http.Header{
				"Authorization": {"Bearer " + validAccessToken},
			},
			expectedStatus:        http.StatusInternalServerError,
			expectedResponse:      "{\"message\":\"wrong token type\",\"status_code\":500}\n",
			expectClaimsInContext: false,
		},
		{
			name: "Missing token type",
			header: http.Header{
				"Authorization": {"Bearer " + generateValidToken(t, issuer, secretKey, "")},
			},
			expectedStatus:        http.StatusInternalServerError,
			expectedResponse:      "{\"message\":\"server error\",\"status_code\":500}\n",
			expectClaimsInContext: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header = tc.header

			recorder := httptest.NewRecorder()

			middleware.ServeHTTP(recorder, req)

			assert.Equal(t, recorder.Code, tc.expectedStatus)

			actualResponse := recorder.Body.String()
			if actualResponse != "" {
				assert.JSONEq(t, actualResponse, tc.expectedResponse)
			}
		})
	}
}
