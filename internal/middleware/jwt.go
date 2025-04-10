package middleware

import (
	"net/http"
	"oneview-be/pkg/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(http.StatusUnauthorized).SendString("Missing or malformed token")
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return []byte(config.Envs.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(http.StatusUnauthorized).SendString("Invalid token")
		}

		c.Locals("user", token)
		return c.Next()
	}
}
