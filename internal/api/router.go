package router

import (
	"github.com/iamrajjoshi/pinguin/internal/api/handlers"
	"github.com/iamrajjoshi/pinguin/internal/monitor"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func New(db *sqlx.DB) *echo.Echo {
	// Create new echo instance
	e := echo.New()

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Initialize services
	monitorService := monitor.NewMonitorService(db)

	// Register handlers
	h := handlers.NewHandler(monitorService)

	api := e.Group("/api")
	{
		monitors := api.Group("/monitors")
		monitors.POST("", h.CreateMonitor)
		monitors.GET("", h.ListMonitors)
		monitors.GET("/:id", h.GetMonitor)
		// TODO(iamrajjoshi): Add checks
	}

	return e
}
