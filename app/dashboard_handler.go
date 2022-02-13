package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func DashboardIndexHandler(c echo.Context) error {
	return r.Inertia(c).Render(http.StatusOK, "Dashboard/Index", nil)
}
