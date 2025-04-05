package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"oneview-be/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func mockToken(userID uint) *jwt.Token {
	return &jwt.Token{
		Claims: jwt.MapClaims{
			"sub": float64(userID),
		},
		Valid: true,
	}
}

func setupMessageTestApp() (*fiber.App, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.Message{})
	app := fiber.New()
	app.Get("/my-messages", func(c *fiber.Ctx) error {
		c.Locals("user", mockToken(1))
		return ListMessages(db)(c)
	})
	app.Post("/send", func(c *fiber.Ctx) error {
		c.Locals("user", mockToken(1))
		return SendMessage(db)(c)
	})
	app.Get("/read/:id", func(c *fiber.Ctx) error {
		return ReadMessage(db)(c)
	})
	return app, db
}

func TestListMessagesEmpty(t *testing.T) {
	app, _ := setupMessageTestApp()
	req := httptest.NewRequest("GET", "/my-messages", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestSendAndReadMessage(t *testing.T) {
	app, db := setupMessageTestApp()
	db.Create(&model.User{ID: 1, Email: "sender@test.com", Password: "", PublicCode: "code1"})
	db.Create(&model.User{ID: 2, Email: "receiver@test.com", Password: "", PublicCode: "code2"})

	body := map[string]string{
		"to_code": "code2",
		"content": "Mensagem secreta!",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/send", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var msg model.Message
	db.First(&msg, "receiver_id = ?", 2)
	if msg.ID == 0 {
		t.Fatal("message not created")
	}

	readReq := httptest.NewRequest("GET", "/read/"+strconv.Itoa(int(msg.ID)), nil)
	readResp, err := app.Test(readReq)
	if err != nil {
		t.Fatal(err)
	}
	if readResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 on read, got %d", readResp.StatusCode)
	}
}
