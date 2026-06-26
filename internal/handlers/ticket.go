package handlers

import (
	"encoding/json"
	"net/http"
	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TicketHandler struct {
	store *store.Store
}

func NewTicketHandler(s *store.Store) *TicketHandler {
	return &TicketHandler{store: s}
}

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

var validTransitions = map[models.TicketStatus]models.TicketStatus{
	models.StatusOpen:       models.StatusInProgress,
	models.StatusInProgress: models.StatusClosed,
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	ticket := &models.Ticket{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusOpen,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.CreateTicket(ticket); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	tickets := h.store.ListTicketsByUser(userID)
	if tickets == nil {
		tickets = []*models.Ticket{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := mux.Vars(r)["id"]

	ticket, err := h.store.GetTicketByID(id)
	if err != nil {
		jsonError(w, "ticket not found", http.StatusNotFound)
		return
	}
	if ticket.UserID != userID {
		jsonError(w, "ticket not found", http.StatusNotFound) // don't leak existence
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := mux.Vars(r)["id"]

	ticket, err := h.store.GetTicketByID(id)
	if err != nil {
		jsonError(w, "ticket not found", http.StatusNotFound)
		return
	}
	if ticket.UserID != userID {
		jsonError(w, "ticket not found", http.StatusNotFound)
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newStatus := models.TicketStatus(req.Status)

	// Validate the requested status is a known value
	switch newStatus {
	case models.StatusOpen, models.StatusInProgress, models.StatusClosed:
		// valid
	default:
		jsonError(w, "invalid status; must be one of: open, in_progress, closed", http.StatusBadRequest)
		return
	}

	// Closed tickets cannot be changed
	if ticket.Status == models.StatusClosed {
		jsonError(w, "closed tickets cannot be reopened", http.StatusUnprocessableEntity)
		return
	}

	// Enforce allowed transition
	allowed, ok := validTransitions[ticket.Status]
	if !ok || allowed != newStatus {
		jsonError(w, "invalid status transition", http.StatusUnprocessableEntity)
		return
	}

	ticket.Status = newStatus
	ticket.UpdatedAt = time.Now()

	if err := h.store.UpdateTicket(ticket); err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}
