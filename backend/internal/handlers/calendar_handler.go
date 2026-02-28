package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thucdx/todovibe/internal/repository"
)

// calendarServicer is the subset of services.CalendarService used by CalendarHandler.
type calendarServicer interface {
	MonthlySummary(ctx context.Context, year, month int) ([]repository.DaySummary, error)
}

// CalendarHandler handles the monthly calendar summary endpoint.
type CalendarHandler struct {
	svc calendarServicer
}

func NewCalendarHandler(svc calendarServicer) *CalendarHandler {
	return &CalendarHandler{svc: svc}
}

// Summary godoc
// GET /api/v1/calendar?year=YYYY&month=MM
func (h *CalendarHandler) Summary(c *gin.Context) {
	now := time.Now()
	year, err := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	month, err := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}
	data, err := h.svc.MonthlySummary(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
