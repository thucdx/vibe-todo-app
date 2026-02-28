package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/repository"
)

// statsServicer is the subset of services.StatsService used by StatsHandler.
type statsServicer interface {
	ChartData(ctx context.Context, view, metric string) ([]repository.ChartPoint, error)
}

// StatsHandler handles the productivity chart data endpoint.
type StatsHandler struct {
	svc statsServicer
}

func NewStatsHandler(svc statsServicer) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// Chart godoc
// GET /api/v1/stats?view=day|week|month&metric=count|points
func (h *StatsHandler) Chart(c *gin.Context) {
	view := c.DefaultQuery("view", "day")
	metric := c.DefaultQuery("metric", "count")
	data, err := h.svc.ChartData(c.Request.Context(), view, metric)
	if err != nil {
		if err == apperrors.ErrBadRequest {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid view or metric parameter"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
