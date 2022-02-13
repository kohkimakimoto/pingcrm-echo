package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ReportsIndexHandler(c echo.Context) error {
	return r.Inertia(c).Render(http.StatusOK, "Reports/Index", nil)
}
