package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
	"time"

	"github.com/gorilla/mux"
)

func TestTicketLifecycle(t *testing.T) {
	s := store.New()
	h := NewTicketHandler(s)

	userID := "user-1"
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)

	// 1. Create a ticket
	createPayload := `{"title":"Test Ticket","description":"Test Description"}`
	req, _ := http.NewRequest(http.MethodPost, "/tickets", bytes.NewBufferString(createPayload))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var ticket models.Ticket
	if err := json.Unmarshal(rr.Body.Bytes(), &ticket); err != nil {
		t.Fatalf("failed to parse created ticket: %v", err)
	}

	if ticket.ID == "" || ticket.Title != "Test Ticket" || ticket.Status != models.StatusOpen || ticket.UserID != userID {
		t.Errorf("created ticket fields are incorrect: %+v", ticket)
	}

	// 2. List tickets
	req, _ = http.NewRequest(http.MethodGet, "/tickets", nil)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var list []*models.Ticket
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to parse ticket list: %v", err)
	}

	if len(list) != 1 || list[0].ID != ticket.ID {
		t.Errorf("expected 1 ticket in list, got: %+v", list)
	}

	// 3. Get ticket by ID (own)
	req, _ = http.NewRequest(http.MethodGet, "/tickets/"+ticket.ID, nil)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": ticket.ID})
	rr = httptest.NewRecorder()
	h.GetByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// 4. Get ticket by ID (other user's ticket - should return 404)
	otherCtx := context.WithValue(context.Background(), middleware.UserIDKey, "other-user")
	req, _ = http.NewRequest(http.MethodGet, "/tickets/"+ticket.ID, nil)
	req = req.WithContext(otherCtx)
	req = mux.SetURLVars(req, map[string]string{"id": ticket.ID})
	rr = httptest.NewRecorder()
	h.GetByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d for other user's ticket, got %d", http.StatusNotFound, rr.Code)
	}

	// 5. Update status: open -> in_progress
	updatePayload := `{"status":"in_progress"}`
	req, _ = http.NewRequest(http.MethodPatch, "/tickets/"+ticket.ID+"/status", bytes.NewBufferString(updatePayload))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": ticket.ID})
	rr = httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var updatedTicket models.Ticket
	json.Unmarshal(rr.Body.Bytes(), &updatedTicket)
	if updatedTicket.Status != models.StatusInProgress {
		t.Errorf("expected status 'in_progress', got '%s'", updatedTicket.Status)
	}

	// 6. Update status: in_progress -> closed
	updatePayload2 := `{"status":"closed"}`
	req, _ = http.NewRequest(http.MethodPatch, "/tickets/"+ticket.ID+"/status", bytes.NewBufferString(updatePayload2))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": ticket.ID})
	rr = httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	json.Unmarshal(rr.Body.Bytes(), &updatedTicket)
	if updatedTicket.Status != models.StatusClosed {
		t.Errorf("expected status 'closed', got '%s'", updatedTicket.Status)
	}

	// 7. Try to update status when closed (should fail)
	updatePayload3 := `{"status":"in_progress"}`
	req, _ = http.NewRequest(http.MethodPatch, "/tickets/"+ticket.ID+"/status", bytes.NewBufferString(updatePayload3))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": ticket.ID})
	rr = httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d (unprocessable entity) when reopening closed ticket, got %d. Body: %s", http.StatusUnprocessableEntity, rr.Code, rr.Body.String())
	}
}

func TestInvalidTransitions(t *testing.T) {
	s := store.New()
	h := NewTicketHandler(s)

	userID := "user-1"
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)

	// Create open ticket
	ticket := &models.Ticket{
		ID:        "t-1",
		Title:     "Test Ticket",
		Status:    models.StatusOpen,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	s.CreateTicket(ticket)

	// Try to jump from open -> closed directly.
	// Wait, is open -> closed directly allowed or forbidden? Let's check if the test fails under current transition logic.
	updatePayload := `{"status":"closed"}`
	req, _ := http.NewRequest(http.MethodPatch, "/tickets/t-1/status", bytes.NewBufferString(updatePayload))
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"id": "t-1"})
	rr := httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	// Current code requires open -> in_progress first. Let's see if we get StatusUnprocessableEntity.
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected %d for direct open->closed, got %d", http.StatusUnprocessableEntity, rr.Code)
	}
}
