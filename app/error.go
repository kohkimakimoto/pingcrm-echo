package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ValidationErrors map[string]string

func (v ValidationErrors) Set(key, message string) ValidationErrors {
	v[key] = message
	return v
}

func ErrorHandler(err error, c echo.Context) {
	hErr := transformToHTTPError(err)
	if hErr.Code >= 500 {
		c.Logger().Error(err)
	}

	if c.Response().Committed {
		return
	}

	var message string
	if msg, ok := hErr.Message.(string); ok {
		message = msg
	} else {
		message = http.StatusText(hErr.Code)
	}

	if err := c.Render(hErr.Code, "error.html", map[string]interface{}{
		"title":   http.StatusText(hErr.Code),
		"code":    hErr.Code,
		"message": message,
	}); err != nil {
		c.Logger().Error(err)
	}
}

func transformToHTTPError(err error) *echo.HTTPError {
	if hErr, ok := err.(*echo.HTTPError); ok {
		return hErr
	}

	code := http.StatusInternalServerError
	hErr := echo.NewHTTPError(code, http.StatusText(code))
	hErr.Internal = err
	return hErr
}
