package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AuthCreateHandler(c echo.Context) error {
	return r.Inertia(c).Render(http.StatusOK, "Auth/Login", nil)
}

func AuthStoreHandler(c echo.Context) error {
	input := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Remember bool   `json:"remember"` // TODO: "remember me" has not been implemented yet.
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}

	// validate
	v := make(ValidationErrors)
	if input.Email == "" {
		v.Set("email", "The email field is required.")
	}
	if input.Password == "" {
		v.Set("password", "The password field is required.")
	}
	if len(v) > 0 {
		if err := r.Session(c).SetErrors(v); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/login")
	}

	user, err := GetUserByEmail(c.Request().Context(), r.DB(c), input.Email)
	if err != nil && !IsErrNoRows(err) {
		return err
	}

	if user == nil || !user.VerifyPassword(input.Password) {
		return r.Inertia(c).Render(http.StatusOK, "Auth/Login", map[string]interface{}{
			"errors": v.Set("email", "These credentials do not match our records."),
		})
	}

	if err := r.Session(c).LoginWithUserId(user.Id); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func AuthDestroyHandler(c echo.Context) error {
	if err := r.Session(c).Logout(); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}
