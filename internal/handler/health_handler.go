package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"aisearch/internal/config"
	"aisearch/pkg/response"
)

type HealthHandler struct {
	cfg config.Config
}

func NewHealthHandler(cfg config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) Check(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"app": h.cfg.AppName,
		"env": h.cfg.Env,
	})
}
