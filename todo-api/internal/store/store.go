package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/todooo1/todo-api/internal/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	mu            sync.RWMutex
	users         map[string]*models.User
	usersByEmail  map[string]*models.User
	tasks         map[string]*models.Task
	categories    map[string]*models.Category
	loginAttempts map[string]*models.LoginAttempt
	resetTokens   map[string]string
}

func New() *Store {
	return &Store{
		users:         make(map[string]*models.User),
		usersByEmail:  make(map[string]*models.User),
		tasks:         make(map[string]*models.Task),
		categories:    make(map[string]*models.Category),
		loginAttempts: make(map[string]*models.LoginAttempt),
		resetTokens:   make(map[string]string),
	}
}

func normEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

// Users

func (s *Store) CreateUser(email, passwordHash string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := normEmail(email)
	if _, ok := s.usersByEmail[key]; ok {
		return nil, ErrAlreadyExists
	}
	u := &models.User{
		ID:           uuid.NewString(),
		Email:        key,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	s.users[u.ID] = u
	s.usersByEmail[u.Email] = u
	return u, nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByEmail[normEmail(email)]
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

func (s *Store) UpdateUser(id, newEmail, newPasswordHash string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	if newEmail != "" {
		newKey := normEmail(newEmail)
		if newKey != u.Email {
			if _, exists := s.usersByEmail[newKey]; exists {
				return nil, ErrAlreadyExists
			}
			delete(s.usersByEmail, u.Email)
			u.Email = newKey
			s.usersByEmail[newKey] = u
		}
	}
	if newPasswordHash != "" {
		u.PasswordHash = newPasswordHash
	}
	return u, nil
}

// Login attempts

func (s *Store) GetLoginAttempt(email string) *models.LoginAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.loginAttempts[normEmail(email)]
	if !ok {
		return nil
	}
	c := *a
	return &c
}

func (s *Store) RecordLoginFailure(email string, window time.Duration, maxFailures int, lockDuration time.Duration) *models.LoginAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := normEmail(email)
	now := time.Now().UTC()
	a, ok := s.loginAttempts[key]
	if !ok || now.Sub(a.FirstFailAt) > window {
		a = &models.LoginAttempt{FirstFailAt: now}
		s.loginAttempts[key] = a
	}
	a.FailedCount++
	if a.FailedCount >= maxFailures {
		a.LockedUntil = now.Add(lockDuration)
	}
	c := *a
	return &c
}

func (s *Store) ClearLoginAttempts(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.loginAttempts, normEmail(email))
}

// Password reset tokens

func (s *Store) SetResetToken(token, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetTokens[token] = userID
}

func (s *Store) ConsumeResetToken(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	uid, ok := s.resetTokens[token]
	if ok {
		delete(s.resetTokens, token)
	}
	return uid, ok
}

// Tasks

func (s *Store) CreateTask(t *models.Task) *models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	s.tasks[t.ID] = t
	return t
}

func (s *Store) GetTask(userID, id string) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok || t.UserID != userID {
		return nil, ErrNotFound
	}
	c := *t
	return &c, nil
}

func (s *Store) UpdateTask(userID, id string, fn func(*models.Task) error) (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok || t.UserID != userID {
		return nil, ErrNotFound
	}
	if err := fn(t); err != nil {
		return nil, err
	}
	t.UpdatedAt = time.Now().UTC()
	c := *t
	return &c, nil
}

func (s *Store) DeleteTask(userID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok || t.UserID != userID {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

func (s *Store) ListTasks(userID string) []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Task, 0)
	for _, t := range s.tasks {
		if t.UserID == userID {
			c := *t
			out = append(out, &c)
		}
	}
	return out
}

// Categories

func (s *Store) CreateCategory(c *models.Category) *models.Category {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	c.CreatedAt = time.Now().UTC()
	s.categories[c.ID] = c
	return c
}

func (s *Store) GetCategory(userID, id string) (*models.Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.categories[id]
	if !ok || c.UserID != userID {
		return nil, ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (s *Store) UpdateCategory(userID, id, name, color string) (*models.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.categories[id]
	if !ok || c.UserID != userID {
		return nil, ErrNotFound
	}
	if name != "" {
		c.Name = name
	}
	if color != "" {
		c.Color = color
	}
	cp := *c
	return &cp, nil
}

func (s *Store) DeleteCategory(userID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.categories[id]
	if !ok || c.UserID != userID {
		return ErrNotFound
	}
	delete(s.categories, id)
	// Unassign from tasks owned by this user
	for _, t := range s.tasks {
		if t.UserID == userID && t.CategoryID == id {
			t.CategoryID = ""
			t.UpdatedAt = time.Now().UTC()
		}
	}
	return nil
}

func (s *Store) ListCategories(userID string) []*models.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Category, 0)
	for _, c := range s.categories {
		if c.UserID == userID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out
}
