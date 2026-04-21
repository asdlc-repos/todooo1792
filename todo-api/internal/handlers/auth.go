package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/todooo1/todo-api/internal/auth"
	"github.com/todooo1/todo-api/internal/httpx"
	"github.com/todooo1/todo-api/internal/store"
	"github.com/todooo1/todo-api/internal/validation"
)

const (
	LoginWindow       = 15 * time.Minute
	LoginMaxFailures  = 5
	LoginLockDuration = 30 * time.Minute
)

type AuthHandler struct {
	Store  *store.Store
	JWT    *auth.JWTManager
	Bcrypt int
}

type credentialsBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      userDTO   `json:"user"`
}

type userDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body credentialsBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email, err := validation.Email(body.Email)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.Password(body.Password); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), h.Bcrypt)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	u, err := h.Store.CreateUser(email, string(hash))
	if err != nil {
		if err == store.ErrAlreadyExists {
			httpx.Error(w, http.StatusConflict, "email already registered")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	token, exp, err := h.JWT.Generate(u.ID, u.Email)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	httpx.JSON(w, http.StatusCreated, tokenResponse{
		Token:     token,
		ExpiresAt: exp,
		User:      userDTO{ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body credentialsBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email, err := validation.Email(body.Email)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if attempt := h.Store.GetLoginAttempt(email); attempt != nil && time.Now().UTC().Before(attempt.LockedUntil) {
		httpx.Error(w, http.StatusTooManyRequests, "account is temporarily locked due to too many failed login attempts")
		return
	}

	u, err := h.Store.GetUserByEmail(email)
	if err != nil {
		h.Store.RecordLoginFailure(email, LoginWindow, LoginMaxFailures, LoginLockDuration)
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		h.Store.RecordLoginFailure(email, LoginWindow, LoginMaxFailures, LoginLockDuration)
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	h.Store.ClearLoginAttempts(email)
	token, exp, err := h.JWT.Generate(u.ID, u.Email)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	httpx.JSON(w, http.StatusOK, tokenResponse{
		Token:     token,
		ExpiresAt: exp,
		User:      userDTO{ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Stateless JWT: client discards the token. We simply acknowledge.
	w.WriteHeader(http.StatusNoContent)
}

type resetRequestBody struct {
	Email    string `json:"email"`
	Token    string `json:"token,omitempty"`
	Password string `json:"password,omitempty"`
}

// PasswordReset initiates a reset (email-less flow) and optionally applies a new password
// when a token and password are supplied.
func (h *AuthHandler) PasswordReset(w http.ResponseWriter, r *http.Request) {
	var body resetRequestBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Token + password: apply reset
	if body.Token != "" && body.Password != "" {
		if err := validation.Password(body.Password); err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		userID, ok := h.Store.ConsumeResetToken(body.Token)
		if !ok {
			httpx.Error(w, http.StatusBadRequest, "invalid or expired reset token")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), h.Bcrypt)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		if _, err := h.Store.UpdateUser(userID, "", string(hash)); err != nil {
			httpx.Error(w, http.StatusBadRequest, "user not found")
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "password updated"})
		return
	}

	// Initiate flow: always respond 202 regardless of existence to prevent enumeration
	email, err := validation.Email(body.Email)
	if err != nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if u, err := h.Store.GetUserByEmail(email); err == nil {
		token := randomToken()
		h.Store.SetResetToken(token, u.ID)
		// In a real system the token would be emailed. Returning it here is not part of
		// the contract; we acknowledge acceptance only.
	}
	w.WriteHeader(http.StatusAccepted)
}

func randomToken() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
