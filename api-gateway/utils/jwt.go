package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
