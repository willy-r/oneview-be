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

	messagesRoutes := r.Group("/messages", middleware.JWTProtected())
	messagesRoutes.Post("/", handler.SendMessage(db))
	messagesRoutes.Get("/my", handler.ListMessages(db))
	messagesRoutes.Get("/:id", handler.ReadMessage(db))

	codeRoutes := r.Group("/code", middleware.JWTProtected())
	codeRoutes.Put("/rotate", handler.RotatePublicCode(db))
	codeRoutes.Get("/my", handler.GetMyPublicCode(db))

	r.Listen(config.Envs.Address)
}
