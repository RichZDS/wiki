package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"aisearch/internal/service"
	"aisearch/pkg/response"
)

type WikiHandler struct {
	service *service.WikiService
}

func NewWikiHandler(service *service.WikiService) *WikiHandler {
	return &WikiHandler{service: service}
}

func (h *WikiHandler) List(c *gin.Context) {
	wikis := h.service.List(c.Request.Context())
	response.Success(c, http.StatusOK, wikis)
}

func (h *WikiHandler) GetByID(c *gin.Context) {
	wiki, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrWikiNotFound) {
			response.Error(c, http.StatusNotFound, "wiki not found")
			return
		}

		response.Error(c, http.StatusInternalServerError, "internal server error")
		return
	}

	response.Success(c, http.StatusOK, wiki)
}
