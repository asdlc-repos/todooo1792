package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/todooo1/todo-api/internal/auth"
	"github.com/todooo1/todo-api/internal/ratelimit"
	"github.com/todooo1/todo-api/internal/server"
	"github.com/todooo1/todo-api/internal/store"
)

func main() {
	port := envDefault("PORT", "9090")
	secret := envDefault("JWT_SECRET", randomHex(32))
	allowOrigins := strings.Split(envDefault("CORS_ALLOW_ORIGINS", "*"), ",")

	s := store.New()
	jwtMgr := auth.NewJWTManager(secret, 24*time.Hour)

	// Public limit: IP-based for auth endpoints (prevents credential stuffing).
	publicLimit := ratelimit.New(100, time.Minute)
	// Authed limit: per-user, 100 requests per minute.
	authedLimit := ratelimit.New(100, time.Minute)

	handler := server.New(server.Config{
		Store:        s,
		JWT:          jwtMgr,
		BcryptCost:   bcrypt.DefaultCost,
		PublicLimit:  publicLimit,
		AuthedLimit:  authedLimit,
		AllowOrigins: allowOrigins,
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Printf("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("todo-api listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	<-idleConnsClosed
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "dev-insecure-default-change-me"
	}
	return hex.EncodeToString(b)
}
