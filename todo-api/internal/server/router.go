package server

import (
	"net/http"
	"strings"

	"github.com/todooo1/todo-api/internal/auth"
	"github.com/todooo1/todo-api/internal/handlers"
	"github.com/todooo1/todo-api/internal/httpx"
	"github.com/todooo1/todo-api/internal/ratelimit"
	"github.com/todooo1/todo-api/internal/store"
)

type Config struct {
	Store        *store.Store
	JWT          *auth.JWTManager
	BcryptCost   int
	PublicLimit  *ratelimit.Limiter
	AuthedLimit  *ratelimit.Limiter
	AllowOrigins []string
}

func New(cfg Config) http.Handler {
	authHandler := &handlers.AuthHandler{Store: cfg.Store, JWT: cfg.JWT, Bcrypt: cfg.BcryptCost}
	accountHandler := &handlers.AccountHandler{Store: cfg.Store, Bcrypt: cfg.BcryptCost}
	taskHandler := &handlers.TaskHandler{Store: cfg.Store}
	categoryHandler := &handlers.CategoryHandler{Store: cfg.Store}

	authMW := auth.Middleware(cfg.JWT)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public auth endpoints (IP-based rate limit)
	publicChain := func(h http.HandlerFunc) http.Handler {
		return cfg.PublicLimit.Middleware()(h)
	}
	mux.Handle("POST /auth/register", publicChain(authHandler.Register))
	mux.Handle("POST /auth/login", publicChain(authHandler.Login))
	mux.Handle("POST /auth/logout", publicChain(authHandler.Logout))
	mux.Handle("POST /auth/password-reset", publicChain(authHandler.PasswordReset))

	// Authenticated endpoints (per-user rate limit)
	authedChain := func(h http.HandlerFunc) http.Handler {
		return authMW(cfg.AuthedLimit.Middleware()(h))
	}

	mux.Handle("PUT /account", authedChain(accountHandler.Update))

	mux.Handle("GET /tasks", authedChain(taskHandler.List))
	mux.Handle("POST /tasks", authedChain(taskHandler.Create))
	mux.Handle("GET /tasks/{id}", authedChain(withID(taskHandler.Get)))
	mux.Handle("PUT /tasks/{id}", authedChain(withID(taskHandler.Update)))
	mux.Handle("DELETE /tasks/{id}", authedChain(withID(taskHandler.Delete)))
	mux.Handle("POST /tasks/{id}/complete", authedChain(withID(taskHandler.Complete)))

	mux.Handle("GET /categories", authedChain(categoryHandler.List))
	mux.Handle("POST /categories", authedChain(categoryHandler.Create))
	mux.Handle("PUT /categories/{id}", authedChain(withID(categoryHandler.Update)))
	mux.Handle("DELETE /categories/{id}", authedChain(withID(categoryHandler.Delete)))

	return corsMiddleware(cfg.AllowOrigins)(recoverMiddleware(loggingMiddleware(mux)))
}

func withID(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			httpx.Error(w, http.StatusBadRequest, "id is required")
			return
		}
		fn(w, r, id)
	}
}

func corsMiddleware(allowed []string) func(http.Handler) http.Handler {
	allowAll := false
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAll = true
		}
		if o != "" {
			allowedSet[o] = struct{}{}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if _, ok := allowedSet[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
