package app

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func ParamInt64(c echo.Context, name string, defaultVal ...int64) int64 {
	if v := getParam(c, name); len(v) > 0 {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}

	if len(defaultVal) > 0 {
		return defaultVal[0]
	}

	return 0
}

func getParam(c echo.Context, name string) string {
	// first: try to get a param from the PATH
	if v := c.Param(name); v != "" {
		return v
	}
	// second: try to get a param from the query string
	if v := c.QueryParam(name); v != "" {
		return v
	}
	// third: try to get a param from the form
	if v := c.FormValue(name); v != "" {
		return v
	}

	return ""
}
