// Package server wires together routing, middleware and handlers.
package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskmanager/internal/auth"
	"taskmanager/internal/config"
	"taskmanager/internal/events"
	"taskmanager/internal/httpx"
	"taskmanager/internal/middleware"
	"taskmanager/internal/tasks"
	"taskmanager/internal/users"
)

// New builds the fully configured HTTP handler.
func New(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	jwtManager := auth.NewManager(cfg.JWTSecret, cfg.JWTTTL)
	hub := events.NewHub()

	userHandler := users.NewHandler(users.NewRepository(pool), jwtManager)
	taskHandler := tasks.NewHandler(tasks.NewRepository(pool), hub)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		// Public auth routes.
		r.Post("/auth/signup", userHandler.Signup)
		r.Post("/auth/login", userHandler.Login)

		// Everything below requires a valid JWT.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtManager))

			r.Get("/auth/me", userHandler.Me)

			r.Route("/tasks", func(r chi.Router) {
				r.Get("/", taskHandler.List)
				r.Post("/", taskHandler.Create)
				r.Get("/stream", taskHandler.Stream)
				r.Get("/{id}", taskHandler.Get)
				r.Patch("/{id}", taskHandler.Update)
				r.Delete("/{id}", taskHandler.Delete)
				r.Get("/{id}/activity", taskHandler.Activity)
			})
		})
	})

	return r
}
