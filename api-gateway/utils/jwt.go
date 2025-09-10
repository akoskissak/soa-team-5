package utils

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/metadata"
)

func GenerateJWT(username string, role string, userId primitive.ObjectID) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"userId":   userId,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(os.Getenv("JWT_SECRET"))
	return token.SignedString(secret)
}

func AuthMetadata(req *http.Request) (metadata.MD, error) {
	publicPaths := map[string]bool{
		"/api/auth/login":    true,
		"/api/auth/register": true,
	}

	if publicPaths[req.URL.Path] {
		return nil, nil
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("missing or invalid Authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return metadata.Pairs(
			"username", claims["username"].(string),
			"userId", claims["userId"].(string),
			"role", claims["role"].(string),
		), nil
	}

	return nil, fmt.Errorf("unable to extract claims from token")
}

func JWTMiddleware(next http.Handler) http.Handler {
	publicPaths := map[string]bool{
		"/api/auth/login":    true,
		"/api/auth/register": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		md, err := AuthMetadata(r)
		if err != nil || md == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
