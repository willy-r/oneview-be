package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"oneview-be/internal/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestAppWithRoutes() (*fiber.App, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.Message{})
	app := fiber.New()
	app.Post("/register", Register(db))
	app.Post("/login", Login(db))
	return app, db
}

func TestRegisterSuccess(t *testing.T) {
	app, _ := setupTestAppWithRoutes()
	payload := AuthRequest{Email: "test@example.com", Password: "StrongPass1!"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	app, _ := setupTestAppWithRoutes()
	payload := AuthRequest{Email: "weak@example.com", Password: "123"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == fiber.StatusCreated {
		t.Errorf("expected error due to weak password, got status %d", resp.StatusCode)
	}
}

func TestLoginSuccess(t *testing.T) {
	app, db := setupTestAppWithRoutes()
	hash, _ := bcrypt.GenerateFromPassword([]byte("Test1234!"), 14)
	db.Create(&model.User{Email: "login@example.com", Password: string(hash), PublicCode: "code123"})

	payload := AuthRequest{Email: "login@example.com", Password: "Test1234!"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}
