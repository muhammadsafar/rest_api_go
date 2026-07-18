package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"restapi/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

func JWTMiddleware(next http.Handler) http.Handler {

	fmt.Println("--------------JWT Middleware------------------")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("++++++++++++Inside  JWT Middleware")

		token, err := r.Cookie("Bearer")
		if err != nil {
			http.Error(w, "Authorization Header missing", http.StatusUnauthorized)
			return
		}
		jwtSecret := os.Getenv("JWT_SECRET")

		parsedToken, err := jwt.Parse(token.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		})

		if err != nil {

			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return

			case errors.Is(err, jwt.ErrTokenNotValidYet):
				http.Error(w, "Token is not valid yet", http.StatusUnauthorized)
				return

			case errors.Is(err, jwt.ErrTokenMalformed):
				http.Error(w, "Malformed token", http.StatusUnauthorized)
				return

			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				return

			case errors.Is(err, jwt.ErrTokenUnverifiable):
				http.Error(w, "Token unverifiable", http.StatusUnauthorized)
				return

			case errors.Is(err, jwt.ErrInvalidKey):
				http.Error(w, "Invalid JWT key", http.StatusInternalServerError)
				return

			default:
				utils.ErrorHandler(err, "JWT validation failed")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		if parsedToken.Valid {
			log.Println("Valid JWT")
		} else {
			http.Error(w, "Invalid login token", http.StatusUnauthorized)
			log.Println("Invalid JWT:", token)
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid login token", http.StatusUnauthorized)
			log.Fatal("Invalid token claims")
			return
		}

		fmt.Println("Token jwt:", token.Value)
		fmt.Println("User:", claims["user"])
		fmt.Println("Exp:", claims["exp"])
		fmt.Println("Role:", claims["role"])

		ctx := context.WithValue(r.Context(), "role", claims["role"])
		ctx = context.WithValue(ctx, "expiresAt", claims["exp"])
		ctx = context.WithValue(ctx, "username", claims["user"])
		ctx = context.WithValue(ctx, "userId", claims["uid"])

		fmt.Println(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
		fmt.Println("Sent response from JWT middleware")
	})
}
