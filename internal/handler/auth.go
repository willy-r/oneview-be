package handler

import (
	"crypto/rand"
	"encoding/hex"
	"oneview-be/internal/model"
	"oneview-be/pkg/config"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func generatePublicCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b) // Ex: "a1f2c3d4"
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#~$%^&*()+|_]{1}`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AuthRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid request")
		}
		if !isStrongPassword(req.Password) {
			return c.Status(400).SendString("Password too weak: use at least 8 characters, with upper/lowercase, a number, and special")
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
		code := generatePublicCode()
		user := model.User{Email: req.Email, Password: string(hash), PublicCode: code}
		if err := db.Create(&user).Error; err != nil {
			return c.Status(409).SendString("User already exists")
		}
		return c.Status(201).JSON(fiber.Map{"public_code": code})
	}
}

func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AuthRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid request")
		}
		var user model.User
		db.First(&user, "email = ?", req.Email)
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return c.Status(401).SendString("Unauthorized")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.ID,
			"exp": time.Now().Add(time.Second * time.Duration(config.Envs.JwtExpirationSeconds)).Unix(),
		})
		signed, _ := token.SignedString([]byte(config.Envs.JwtSecret))
		return c.JSON(fiber.Map{"token": signed})
	}
}

func RotatePublicCode(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))

		newCode := generatePublicCode()
		if err := db.Model(&model.User{}).Where("id = ?", userID).Update("public_code", newCode).Error; err != nil {
			return c.Status(500).SendString("Failed to rotate public code")
		}
		return c.JSON(fiber.Map{"new_public_code": newCode})
	}
}

func GetMyPublicCode(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))

		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			return c.Status(404).SendString("User not found")
		}
		return c.JSON(fiber.Map{"public_code": user.PublicCode})
	}
}
