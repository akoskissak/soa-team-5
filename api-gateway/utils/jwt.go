package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	stakeproto "stakeholders-service/proto/stakeholders"
)

type contextKey string

const (
	UserIDKey   contextKey = "userId"
	UsernameKey contextKey = "username"
	RoleKey     contextKey = "role"
)

func AuthMetadata(ctx context.Context) (metadata.MD, error) {
	userId, ok1 := ctx.Value(UserIDKey).(string)
	username, ok2 := ctx.Value(UsernameKey).(string)
	role, ok3 := ctx.Value(RoleKey).(string)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("missing user data in context")
	}

	return metadata.Pairs(
		"userId", userId,
		"username", username,
		"role", role,
	), nil
}

func JWTMiddleware(next http.Handler, stakeholdersClient stakeproto.StakeholdersServiceClient) http.Handler {
	publicPaths := map[string]bool{
		"/api/auth/login":    true,
		"/api/auth/register": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		validateReq := &stakeproto.ValidateTokenRequest{Token: tokenStr}

		validateRes, err := stakeholdersClient.ValidateToken(r.Context(), validateReq)

		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		if !validateRes.IsValid {
			http.Error(w, "Unauthorized: Invalid Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, validateRes.UserId)
		ctx = context.WithValue(ctx, UsernameKey, validateRes.Username)
		ctx = context.WithValue(ctx, RoleKey, validateRes.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
