package handlers

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/todooo1/todo-api/internal/auth"
	"github.com/todooo1/todo-api/internal/httpx"
	"github.com/todooo1/todo-api/internal/models"
	"github.com/todooo1/todo-api/internal/store"
	"github.com/todooo1/todo-api/internal/validation"
)

type TaskHandler struct {
	Store *store.Store
}

type taskBody struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	CategoryID  *string    `json:"categoryId,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	Completed   *bool      `json:"completed,omitempty"`
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	tasks := h.Store.ListTasks(userID)

	q := r.URL.Query()
	category := strings.TrimSpace(q.Get("category"))
	status := strings.ToLower(strings.TrimSpace(q.Get("status")))
	dateRange := strings.ToLower(strings.TrimSpace(q.Get("dateRange")))
	search := strings.ToLower(strings.TrimSpace(q.Get("q")))

	filtered := tasks[:0]
	now := time.Now().UTC()
	for _, t := range tasks {
		if category != "" && t.CategoryID != category {
			continue
		}
		switch status {
		case "completed":
			if !t.Completed {
				continue
			}
		case "incomplete", "active", "pending":
			if t.Completed {
				continue
			}
		case "overdue":
			if t.Completed || t.DueDate == nil || !t.DueDate.Before(now) {
				continue
			}
		}
		if dateRange != "" && !matchDateRange(t, dateRange, now) {
			continue
		}
		if search != "" {
			hay := strings.ToLower(t.Title + " " + t.Description)
			if !strings.Contains(hay, search) {
				continue
			}
		}
		filtered = append(filtered, t)
	}

	// Sort chronologically by due date ascending, undated last, then by createdAt desc.
	sort.SliceStable(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		if a.DueDate == nil && b.DueDate == nil {
			return a.CreatedAt.After(b.CreatedAt)
		}
		if a.DueDate == nil {
			return false
		}
		if b.DueDate == nil {
			return true
		}
		if !a.DueDate.Equal(*b.DueDate) {
			return a.DueDate.Before(*b.DueDate)
		}
		return a.CreatedAt.After(b.CreatedAt)
	})

	httpx.JSON(w, http.StatusOK, filtered)
}

func matchDateRange(t *models.Task, rng string, now time.Time) bool {
	if t.DueDate == nil {
		return rng == "none" || rng == "undated"
	}
	due := t.DueDate.UTC()
	y, m, d := now.Date()
	startOfDay := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)
	switch rng {
	case "today":
		return !due.Before(startOfDay) && due.Before(endOfDay)
	case "week":
		endOfWeek := startOfDay.Add(7 * 24 * time.Hour)
		return !due.Before(startOfDay) && due.Before(endOfWeek)
	case "month":
		endOfMonth := startOfDay.AddDate(0, 1, 0)
		return !due.Before(startOfDay) && due.Before(endOfMonth)
	case "overdue":
		return due.Before(startOfDay) && !t.Completed
	case "upcoming", "future":
		return !due.Before(endOfDay)
	}
	return true
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	var body taskBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == nil {
		httpx.Error(w, http.StatusBadRequest, "title is required")
		return
	}
	title, err := validation.Title(*body.Title)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	var desc string
	if body.Description != nil {
		d, err := validation.Description(*body.Description)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		desc = d
	}
	var categoryID string
	if body.CategoryID != nil && *body.CategoryID != "" {
		if _, err := h.Store.GetCategory(userID, *body.CategoryID); err != nil {
			httpx.Error(w, http.StatusBadRequest, "category not found")
			return
		}
		categoryID = *body.CategoryID
	}
	t := &models.Task{
		UserID:      userID,
		Title:       title,
		Description: desc,
		CategoryID:  categoryID,
		DueDate:     body.DueDate,
	}
	if body.Completed != nil && *body.Completed {
		now := time.Now().UTC()
		t.Completed = true
		t.CompletedAt = &now
	}
	created := h.Store.CreateTask(t)
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request, id string) {
	userID := auth.UserID(r)
	t, err := h.Store.GetTask(userID, id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "task not found")
		return
	}
	httpx.JSON(w, http.StatusOK, t)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request, id string) {
	userID := auth.UserID(r)
	var body taskBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.Store.UpdateTask(userID, id, func(t *models.Task) error {
		if body.Title != nil {
			title, err := validation.Title(*body.Title)
			if err != nil {
				return err
			}
			t.Title = title
		}
		if body.Description != nil {
			d, err := validation.Description(*body.Description)
			if err != nil {
				return err
			}
			t.Description = d
		}
		if body.CategoryID != nil {
			if *body.CategoryID == "" {
				t.CategoryID = ""
			} else {
				if _, err := h.Store.GetCategory(userID, *body.CategoryID); err != nil {
					return err
				}
				t.CategoryID = *body.CategoryID
			}
		}
		if body.DueDate != nil {
			d := *body.DueDate
			t.DueDate = &d
		}
		if body.Completed != nil {
			if *body.Completed && !t.Completed {
				now := time.Now().UTC()
				t.CompletedAt = &now
			} else if !*body.Completed {
				t.CompletedAt = nil
			}
			t.Completed = *body.Completed
		}
		return nil
	})
	if err != nil {
		if err == store.ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "task not found")
			return
		}
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, updated)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request, id string) {
	userID := auth.UserID(r)
	if err := h.Store.DeleteTask(userID, id); err != nil {
		httpx.Error(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) Complete(w http.ResponseWriter, r *http.Request, id string) {
	userID := auth.UserID(r)
	updated, err := h.Store.UpdateTask(userID, id, func(t *models.Task) error {
		if !t.Completed {
			now := time.Now().UTC()
			t.Completed = true
			t.CompletedAt = &now
		} else {
			// Toggle back to incomplete if already complete
			t.Completed = false
			t.CompletedAt = nil
		}
		return nil
	})
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "task not found")
		return
	}
	httpx.JSON(w, http.StatusOK, updated)
}
