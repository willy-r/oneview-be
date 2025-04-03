package handler

import (
	"oneview-be/internal/model"
	"oneview-be/pkg/config"
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

func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AuthRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid request")
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
		user := model.User{Email: req.Email, Password: string(hash)}
		if err := db.Create(&user).Error; err != nil {
			return c.Status(400).SendString("User already exists")
		}
		return c.SendStatus(201)
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
