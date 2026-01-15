package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Context keys
type key string

const (
	UserIDContextKey key = "userID"
	UserContextKey   key = "user"
)

// JWT Secret key - loaded from environment or use default for development
var JwtKey []byte

func init() {
	// Load JWT secret from environment variable
	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		JwtKey = []byte(secret)
	} else {
		JwtKey = []byte("super_secret_key_change_in_production")
		fmt.Println("WARNING: Using default JWT secret. Set JWT_SECRET environment variable in production.")
	}
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID string, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID, // Note: "user_id" not "userID"
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtKey)
}

// JWTMiddleware validates JWT tokens and adds user info to context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public routes
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Public routes - allow GET /posts (feed) without auth
		publicRoutes := map[string][]string{
			"/login":      {"POST"},
			"/signup":     {"POST"},
			"/posts":      {"GET"}, // GET /posts is public feed
			"/posts/feed": {"GET"}, // Also allow /posts/feed
		}

		if methods, exists := publicRoutes[r.URL.Path]; exists {
			for _, method := range methods {
				if r.Method == method {
					fmt.Printf("DEBUG: Allowing public access to %s %s\n", r.Method, r.URL.Path)
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		fmt.Printf("DEBUG Middleware: Checking auth for %s %s\n", r.Method, r.URL.Path)

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		fmt.Printf("DEBUG: Authorization header: %s\n", authHeader)

		if authHeader == "" {
			fmt.Println("DEBUG: No Authorization header found")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Missing Authorization header",
			})
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			fmt.Printf("DEBUG: Invalid Authorization format. Parts: %v\n", parts)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid Authorization format. Use: Bearer <token>",
			})
			return
		}

		if parts[0] != "Bearer" {
			fmt.Printf("DEBUG: Missing Bearer prefix. Got: %s\n", parts[0])
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authorization must start with Bearer",
			})
			return
		}

		tokenStr := parts[1]
		fmt.Printf("DEBUG: Token string length: %d\n", len(tokenStr))
		fmt.Printf("DEBUG: Token (first 30 chars): %s...\n", tokenStr[:min(30, len(tokenStr))])

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JwtKey, nil
		})

		if err != nil {
			fmt.Printf("DEBUG: Token parse error: %v\n", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Invalid token: %v", err),
			})
			return
		}

		if !token.Valid {
			fmt.Println("DEBUG: Token is not valid")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid token",
			})
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("DEBUG: Failed to extract claims")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid token claims",
			})
			return
		}

		// Print all claims for debugging
		fmt.Printf("DEBUG: Token claims: %+v\n", claims)

		// Check if token is expired
		if exp, ok := claims["exp"].(float64); ok {
			expiryTime := time.Unix(int64(exp), 0)
			if time.Now().After(expiryTime) {
				fmt.Printf("DEBUG: Token expired at %v\n", expiryTime)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Token expired",
				})
				return
			}
		}

		// Extract userID from claims
		userIDInterface, exists := claims["user_id"]
		if !exists {
			fmt.Println("DEBUG: user_id claim not found in token")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "User ID not found in token",
			})
			return
		}

		userID, ok := userIDInterface.(string)
		if !ok {
			fmt.Printf("DEBUG: user_id is not string. Type: %T, Value: %v\n", userIDInterface, userIDInterface)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid user ID in token",
			})
			return
		}

		if userID == "" {
			fmt.Println("DEBUG: user_id is empty string")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Empty user ID in token",
			})
			return
		}

		fmt.Printf("DEBUG Middleware: Authenticated userID: %s\n", userID)

		// Store userID in context
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		// Also store full claims if needed
		ctx = context.WithValue(ctx, UserContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts userID from request context
func GetUserIDFromContext(r *http.Request) (string, error) {
	userIDValue := r.Context().Value(UserIDContextKey)
	if userIDValue == nil {
		return "", fmt.Errorf("userID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		return "", fmt.Errorf("invalid userID type in context")
	}

	return userID, nil
}

// ValidateToken parses and validates a JWT token string
func ValidateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetUserIDFromToken extracts userID from a JWT token string
func GetUserIDFromToken(tokenStr string) (string, error) {
	claims, err := ValidateToken(tokenStr)
	if err != nil {
		return "", err
	}

	userIDInterface, exists := claims["user_id"]
	if !exists {
		return "", fmt.Errorf("user_id not found in token")
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		return "", fmt.Errorf("invalid user_id type in token")
	}

	return userID, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
