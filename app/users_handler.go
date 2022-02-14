package app

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func UsersIndexHandler(c echo.Context) error {
	filters := &UserFilters{}
	if err := c.Bind(filters); err != nil {
		return err
	}

	filters.AccountId = r.User(c).AccountId

	users, err := ListUsersByFilters(c.Request().Context(), r.DB(c), filters)
	if err != nil {
		return err
	}

	return r.Inertia(c).Render(http.StatusOK, "Users/Index", map[string]interface{}{
		"filters": filters,
		"users": users.ToMap(
			"id",
			"name",
			"email",
			"owner",
			"photo",
			"deleted_at",
		),
	})
}

func UsersCreateHandler(c echo.Context) error {
	return r.Inertia(c).Render(http.StatusOK, "Users/Create", nil)
}

func UsersStoreHandler(c echo.Context) error {
	input := &struct {
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Email     string  `json:"email"`
		Password  *string `json:"password"`
		Owner     bool    `json:"owner"`
		// Photo is not implemented yet
	}{}
	if err := c.Bind(input); err != nil {
		return err
	}

	va := NewValidator()
	va.Required("first_name", input.FirstName, "The first name field is required.")
	va.Required("last_name", input.LastName, "The last name field is required.")
	va.Required("email", input.Email, "The email field is required.")

	_, err := GetUserByEmail(c.Request().Context(), r.DB(c), input.Email)
	if err == nil {
		va.SetError("email", input.Email, "The email has already been taken.")
	} else if !IsErrNoRows(err) {
		return err
	}

	if va.HasErrors() {
		if err := r.Session(c).SetErrors(va.ErrorMessageMap()); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/users/create")
	}

	user := NewUser()
	user.AccountId = r.User(c).AccountId
	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.Email = input.Email
	user.Owner = input.Owner
	if input.Password != nil {
		user.SetPlainPassword(*input.Password)
	}

	if err := CreateUser(c.Request().Context(), r.DB(c), user); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("User created."); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/users")
}

func UsersEditHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	user, err := GetUserById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}
	return r.Inertia(c).Render(http.StatusOK, "Users/Edit", map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.Id,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"owner":      user.Owner,
			"photo":      nil, // not implemented yet
			"deleted_at": user.DeletedAt,
		},
	})
}

func UsersUpdateHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	user, err := GetUserById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}

	if r.IsDemo(c) && user.IsDemoUser() {
		if err := r.Session(c).SetFlashError("Updating the demo user is not allowed."); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, fmt.Sprintf("/users/%d/edit", user.Id))
	}

	input := &struct {
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Email     string  `json:"email"`
		Password  *string `json:"password"`
		Owner     bool    `json:"owner"`
		// Photo is not implemented yet
	}{}
	if err := c.Bind(input); err != nil {
		return err
	}

	va := NewValidator()
	va.Required("first_name", input.FirstName, "The first name field is required.")
	va.Required("last_name", input.LastName, "The last name field is required.")
	va.Required("email", input.Email, "The email field is required.")

	user2, err := GetUserByEmail(c.Request().Context(), r.DB(c), input.Email)
	if err == nil && user2.Id != user.Id {
		va.SetError("email", input.Email, "The email has already been taken.")
	} else if err != nil && !IsErrNoRows(err) {
		return err
	}

	if va.HasErrors() {
		if err := r.Session(c).SetErrors(va.ErrorMessageMap()); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, fmt.Sprintf("/users/%d/edit", user.Id))
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.Email = input.Email
	user.Owner = input.Owner
	if input.Password != nil {
		user.SetPlainPassword(*input.Password)
	}

	if err := UpdateUser(c.Request().Context(), r.DB(c), user); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("User updated."); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/users/%d/edit", user.Id))
}

func UsersDestroyHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	user, err := GetUserById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}

	if r.IsDemo(c) && user.IsDemoUser() {
		if err := r.Session(c).SetFlashError("Deleting the demo user is not allowed."); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, fmt.Sprintf("/users/%d/edit", user.Id))
	}

	if err := SoftDeleteUser(c.Request().Context(), r.DB(c), id); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("User deleted."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/users/%v/edit", id))
}

func UsersRestoreHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := RestoreUser(c.Request().Context(), r.DB(c), id); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("User restored."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/users/%v/edit", id))
}
