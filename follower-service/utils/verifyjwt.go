package utils

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GetClaimsFromJWT(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", "", errors.New("missing or invalid Authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("could not parse JWT claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", "", errors.New("username claim not found")
	}
	userId, ok := claims["userId"].(string)
	if !ok {
		return "", "", errors.New("userId claim not found")
	}
	return username, userId, nil
}

func GetClaimsFromContext(ctx context.Context) (string, string, error) {
	// Dohvati metadatu (HTTP headere) iz gRPC konteksta
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// Dohvati vrijednost "authorization" headera
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", "", status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// Izdvoji "Bearer" token iz headera
	bearerToken := values[0]
	tokenStr := strings.TrimPrefix(bearerToken, "Bearer ")
	if tokenStr == bearerToken {
		return "", "", status.Error(codes.Unauthenticated, "authorization token is not in the 'Bearer <token>' format")
	}

	// Provjeri i parsiraj JWT token
	claims, err := VerifyJWTString(tokenStr)
	if err != nil {
		return "", "", status.Error(codes.Unauthenticated, err.Error())
	}

	// Izvadi username i userId iz claims-a
	username, ok := claims["username"].(string)
	if !ok {
		return "", "", status.Error(codes.Internal, "username claim not found or is not a string")
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		return "", "", status.Error(codes.Internal, "userId claim not found or is not a string")
	}

	return username, userId, nil
}

func VerifyJWTString(tokenStr string) (jwt.MapClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not parse JWT claims")
	}

	return claims, nil
}
