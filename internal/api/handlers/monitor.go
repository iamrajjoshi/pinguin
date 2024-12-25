package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	service "github.com/iamrajjoshi/pinguin/internal/monitor"
	store "github.com/iamrajjoshi/pinguin/internal/store/models"
)

type Handler struct {
	monitorService *service.PostgresMonitorService
}

func NewHandler(ms *service.PostgresMonitorService) *Handler {
	return &Handler{
		monitorService: ms,
	}
}

type CreateMonitorRequest struct {
	URL      string `json:"url" validate:"required,url"`
	Name     string `json:"name" validate:"required"`
	Interval int    `json:"interval" validate:"required,min=30"` // seconds
}

type UpdateMonitorRequest = CreateMonitorRequest

func (h *Handler) CreateMonitor(c echo.Context) error {
	var req CreateMonitorRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	monitor := &store.Monitor{
		URL:      req.URL,
		Name:     req.Name,
		Interval: req.Interval,
		Enabled:  true,
	}

	if err := h.monitorService.Create(c.Request().Context(), monitor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// TODO: Schedule the monitor

	return c.JSON(http.StatusCreated, monitor)
}

func (h *Handler) ListMonitors(c echo.Context) error {
	monitors, err := h.monitorService.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, monitors)
}

func (h *Handler) GetMonitor(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid monitor id")
	}

	monitor, err := h.monitorService.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, monitor)
}

func (h *Handler) UpdateMonitor(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid monitor id")
	}

	var req UpdateMonitorRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	monitor := &store.Monitor{
		ID:       id,
		URL:      req.URL,
		Name:     req.Name,
		Interval: req.Interval,
	}

	if err := h.monitorService.Update(c.Request().Context(), monitor); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// TODO: Update the monitor in the scheduler

	return c.JSON(http.StatusOK, monitor)
}
