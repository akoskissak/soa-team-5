package utils

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func GetClaimsFromContext2Args(ctx context.Context) (jwt.MapClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata not provided")
	}

	username := md.Get("username")
	userId := md.Get("userId")
	role := md.Get("role")

	if len(username) == 0 || len(userId) == 0 || len(role) == 0 {
		return nil, status.Error(codes.Unauthenticated, "claims missing in metadata")
	}

	claims := jwt.MapClaims{
		"username": username[0],
		"userId":   userId[0],
		"role":     role[0],
	}

	return claims, nil
}