package handlers

import (
	"net/http"

	"github.com/todooo1/todo-api/internal/auth"
	"github.com/todooo1/todo-api/internal/httpx"
	"github.com/todooo1/todo-api/internal/models"
	"github.com/todooo1/todo-api/internal/store"
	"github.com/todooo1/todo-api/internal/validation"
)

type CategoryHandler struct {
	Store *store.Store
}

type categoryBody struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	cats := h.Store.ListCategories(userID)
	httpx.JSON(w, http.StatusOK, cats)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	var body categoryBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	name, err := validation.CategoryName(body.Name)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	c := &models.Category{
		UserID: userID,
		Name:   name,
		Color:  validation.Sanitize(body.Color),
	}
	created := h.Store.CreateCategory(c)
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	userID := auth.UserID(r)
	var body categoryBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var name string
	if body.Name != "" {
		n, err := validation.CategoryName(body.Name)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		name = n
	}
	updated, err := h.Store.UpdateCategory(userID, id, name, validation.Sanitize(body.Color))
	if err != nil {
		if err == store.ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "category not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "failed to update category")
		return
	}
	httpx.JSON(w, http.StatusOK, updated)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	userID := auth.UserID(r)
	if err := h.Store.DeleteCategory(userID, id); err != nil {
		httpx.Error(w, http.StatusNotFound, "category not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
