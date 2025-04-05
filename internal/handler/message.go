package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"oneview-be/internal/model"
	"oneview-be/pkg/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var secretKey = []byte(config.Envs.AESKey)

type MessageRequest struct {
	ToCode  string `json:"to_code"`
	Content string `json:"content"`
}

func encrypt(text string) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(cryptoText string) (string, error) {
	ciphertext, _ := base64.StdEncoding.DecodeString(cryptoText)
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}

func SendMessage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		senderID := uint(claims["sub"].(float64))

		var req MessageRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid body")
		}

		var receiver model.User
		if err := db.First(&receiver, "public_code = ?", req.ToCode).Error; err != nil {
			return c.Status(404).SendString("Receiver not found")
		}

		var sender model.User
		if err := db.First(&sender, senderID).Error; err != nil {
			return c.Status(404).SendString("Sender not found")
		}

		encrypted, err := encrypt(req.Content)
		if err != nil {
			return c.Status(500).SendString("Encryption failed")
		}

		msg := model.Message{
			SenderID:       senderID,
			ReceiverID:     receiver.ID,
			SenderCode:     sender.PublicCode,
			ReceiverCode:   receiver.PublicCode,
			Content:        encrypted,
			CreatedAt:      time.Now(),
			ExpirationTime: time.Now().Add(time.Second * time.Duration(config.Envs.MessagesExpirationSeconds)),
		}
		db.Create(&msg)

		notifyUser(receiver.ID, "new_message")

		return c.Status(201).JSON(msg.ID)
	}
}

func ReadMessage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var msg model.Message
		if err := db.First(&msg, id).Error; err != nil {
			return c.Status(404).SendString("Not found")
		}
		if msg.ReadAt != nil {
			return c.Status(410).SendString("Message already read")
		}
		now := time.Now()
		if msg.ExpirationTime.Before(now) {
			db.Delete(&msg)
			return c.Status(410).SendString("Message expired")
		}
		db.Model(&msg).Update("read_at", &now)
		db.Delete(&msg)

		decrypted, err := decrypt(msg.Content)
		if err != nil {
			return c.Status(500).SendString("Decryption failed")
		}

		notifyUser(msg.SenderID, "message_read")

		return c.JSON(fiber.Map{"message": decrypted, "sender_code": msg.SenderCode})
	}
}

func ListMessages(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))

		var messagesToDelete []model.Message
		db.Where("receiver_id = ? AND read_at IS NULL", userID).Where("expiration_time < ?", time.Now()).Find(&messagesToDelete)
		for _, m := range messagesToDelete {
			db.Delete(&m)
		}

		var messages []model.Message
		db.Where("receiver_id = ? AND read_at IS NULL", userID).Order("created_at desc").Find(&messages)

		var result []fiber.Map = make([]fiber.Map, 0, len(messages))
		for _, m := range messages {
			result = append(result, fiber.Map{
				"id":         m.ID,
				"created_at": m.CreatedAt,
				"read_at":    m.ReadAt,
			})
		}

		return c.JSON(result)
	}
}

var clients = make(map[uint]*websocket.Conn)
var mu sync.Mutex

func WebSocketHandler(c *websocket.Conn) {
	userIDStr := c.Params("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Println("Invalid userID")
		c.Close()
		return
	}

	mu.Lock()
	clients[uint(userID)] = c
	mu.Unlock()
	log.Printf("WebSocket connected: user %d\n", userID)

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}

	mu.Lock()
	delete(clients, uint(userID))
	mu.Unlock()
	log.Printf("WebSocket disconnected: user %d\n", userID)
}

func notifyUser(userID uint, message string) {
	mu.Lock()
	defer mu.Unlock()
	if conn, ok := clients[userID]; ok {
		conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
