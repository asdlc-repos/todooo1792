package handlers

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/todooo1/todo-api/internal/auth"
	"github.com/todooo1/todo-api/internal/httpx"
	"github.com/todooo1/todo-api/internal/store"
	"github.com/todooo1/todo-api/internal/validation"
)

type AccountHandler struct {
	Store  *store.Store
	Bcrypt int
}

type updateAccountBody struct {
	Email           string `json:"email,omitempty"`
	CurrentPassword string `json:"currentPassword,omitempty"`
	NewPassword     string `json:"newPassword,omitempty"`
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	var body updateAccountBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	u, err := h.Store.GetUserByID(userID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "user not found")
		return
	}

	if body.NewPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.CurrentPassword)); err != nil {
			httpx.Error(w, http.StatusUnauthorized, "current password is incorrect")
			return
		}
		if err := validation.Password(body.NewPassword); err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	var newEmail string
	if body.Email != "" {
		e, err := validation.Email(body.Email)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		newEmail = e
	}

	var newHash string
	if body.NewPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), h.Bcrypt)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		newHash = string(hash)
	}

	updated, err := h.Store.UpdateUser(userID, newEmail, newHash)
	if err != nil {
		if err == store.ErrAlreadyExists {
			httpx.Error(w, http.StatusConflict, "email already in use")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "failed to update account")
		return
	}
	httpx.JSON(w, http.StatusOK, userDTO{ID: updated.ID, Email: updated.Email, CreatedAt: updated.CreatedAt})
}
