package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/vpramatarov/pdf-tools/internal/api/handlers"
)

func New(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(120 * time.Second))

	// Routes
	r.Get("/", h.Home)
	r.Post("/compress", h.Compress)
	r.Get("/download/{filename}", h.Download)
	r.Post("/convert-word", h.ConvertToWord)

	return r
}
