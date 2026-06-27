package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"ticket-system/internal/store"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	s := store.New()
	h := NewAuthHandler(s)

	// Test registration
	regPayload := `{"email":"test@example.com","password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regPayload))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var regResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("failed to parse registration response: %v", err)
	}

	if _, ok := regResp["id"]; !ok {
		t.Errorf("expected id in response, got none")
	}
	if email, ok := regResp["email"].(string); !ok || email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%v'", regResp["email"])
	}

	// Test conflict registration
	req, _ = http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regPayload))
	rr = httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusConflict {
		t.Errorf("expected conflict status %d, got %d", http.StatusConflict, rr.Code)
	}

	// Test invalid email/password requirements
	invalidPayloads := []string{
		`{"email":"","password":"password123"}`,
		`{"email":"test@example.com","password":""}`,
		`{"email":"test@example.com","password":"123"}`, // too short
	}
	for _, p := range invalidPayloads {
		req, _ = http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(p))
		rr = httptest.NewRecorder()
		h.Register(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected bad request status %d for payload %s, got %d", http.StatusBadRequest, p, rr.Code)
		}
	}

	// Test login
	loginPayload := `{"email":"test@example.com","password":"password123"}`
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginPayload))
	rr = httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected login status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var loginResp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	token := loginResp["token"]
	if token == "" {
		t.Errorf("expected jwt token, got empty string")
	}

	// Test invalid login credentials
	badLoginPayload := `{"email":"test@example.com","password":"wrongpassword"}`
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(badLoginPayload))
	rr = httptest.NewRecorder()
	h.Login(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected unauthorized status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	nonExistentLoginPayload := `{"email":"nobody@example.com","password":"password123"}`
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(nonExistentLoginPayload))
	rr = httptest.NewRecorder()
	h.Login(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected unauthorized status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestEmailNormalization(t *testing.T) {
	s := store.New()
	h := NewAuthHandler(s)

	// Register with mixed case and spaces
	regPayload := `{"email":"  User@Example.Com  ","password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regPayload))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var regResp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &regResp)
	if regResp["email"] != "user@example.com" {
		t.Errorf("expected normalized email 'user@example.com', got '%v'", regResp["email"])
	}

	// Login with different casing/spacing
	loginPayload := `{"email":" USER@example.com ","password":"password123"}`
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginPayload))
	rr = httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected login status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}
