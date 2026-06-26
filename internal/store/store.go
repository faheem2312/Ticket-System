package store

import (
	"errors"
	"sync"
	"ticket-system/internal/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	mu      sync.RWMutex
	users   map[string]*models.User   // id -> user
	byEmail map[string]*models.User   // email -> user
	tickets map[string]*models.Ticket // id -> ticket
}

func New() *Store {
	return &Store{
		users:   make(map[string]*models.User),
		byEmail: make(map[string]*models.User),
		tickets: make(map[string]*models.Ticket),
	}
}

// --- User methods ---

func (s *Store) CreateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byEmail[u.Email]; exists {
		return ErrAlreadyExists
	}
	s.users[u.ID] = u
	s.byEmail[u.Email] = u
	return nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Store) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

// --- Ticket methods ---

func (s *Store) CreateTicket(t *models.Ticket) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[t.ID] = t
	return nil
}

func (s *Store) GetTicketByID(id string) (*models.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *Store) ListTicketsByUser(userID string) []*models.Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*models.Ticket
	for _, t := range s.tickets {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) UpdateTicket(t *models.Ticket) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tickets[t.ID]; !ok {
		return ErrNotFound
	}
	s.tickets[t.ID] = t
	return nil
}
