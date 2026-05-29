package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oalpha/internal/auth"
)

func TestAuthMiddlewareRequiresValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := auth.NewAuthService(nil, "test-secret", time.Hour)

	router := gin.New()
	router.GET("/protected", AuthMiddleware(authService), func(c *gin.Context) {
		value, exists := c.Get(contextUserIDKey)
		if !exists {
			t.Fatal("expected user ID in context")
		}
		c.JSON(http.StatusOK, gin.H{"user_id": value})
	})

	missingTokenReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	missingTokenResp := httptest.NewRecorder()
	router.ServeHTTP(missingTokenResp, missingTokenReq)
	if missingTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token to return 401, got %d", missingTokenResp.Code)
	}

	validToken := signedTestToken(t, "test-secret", 42)
	validReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	validReq.Header.Set("Authorization", "Bearer "+validToken)
	validResp := httptest.NewRecorder()
	router.ServeHTTP(validResp, validReq)
	if validResp.Code != http.StatusOK {
		t.Fatalf("expected valid token to return 200, got %d", validResp.Code)
	}
}

func signedTestToken(t *testing.T, secret string, userID int64) string {
	t.Helper()

	claims := auth.Claims{
		UserID:   userID,
		Username: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
