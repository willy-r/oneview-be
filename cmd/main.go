package main

import (
	"oneview-be/internal/handler"
	"oneview-be/internal/middleware"
	"oneview-be/pkg/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	db := config.InitDatabase()
	r := fiber.New()

	r.Post("/register", handler.Register(db))
	r.Post("/login", handler.Login(db))

	r.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	r.Get("/ws/:userID", websocket.New(handler.WebSocketHandler))

	api := r.Group("/messages", middleware.JWTProtected())
	api.Post("/", handler.SendMessage(db))
	api.Get("/:id", handler.ReadMessage(db))

	r.Listen(config.Envs.Address)
}
