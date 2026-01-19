package handlers

import "github.com/vpramatarov/pdf-tools/internal/config"

type Handler struct {
	Cfg *config.Config
}

func New(cfg *config.Config) *Handler {
	return &Handler{
		Cfg: cfg,
	}
}
