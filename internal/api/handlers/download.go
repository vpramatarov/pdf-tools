package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	cleanPath := filepath.Join(h.Cfg.UploadDir, filepath.Base(filename))

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, cleanPath)
}
