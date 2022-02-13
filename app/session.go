package app

import (
	"encoding/gob"
	"errors"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func SessionMiddleware(secret []byte) echo.MiddlewareFunc {
	return session.Middleware(sessions.NewCookieStore(secret))
}

type SessionWrapper struct {
	c    echo.Context
	s    *sessions.Session
	user *User
}

func NewSessionWrapper(c echo.Context) (*SessionWrapper, error) {
	sess, err := session.Get("session", c)
	if err != nil && err.Error() != securecookie.ErrMacInvalid.Error() {
		return nil, err
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	return &SessionWrapper{
		c: c,
		s: sess,
	}, nil
}

func (s *SessionWrapper) Set(key string, value interface{}) error {
	s.s.Values[key] = value
	return s.s.Save(s.c.Request(), s.c.Response())
}

func (s *SessionWrapper) LoginWithUserId(id int64) error {
	return s.Set("user_id", id)
}

func (s *SessionWrapper) Logout() error {
	s.s.Options.MaxAge = -1
	return s.s.Save(s.c.Request(), s.c.Response())
}

func (s *SessionWrapper) UserId() (int64, error) {
	if id, ok := s.s.Values["user_id"]; ok {
		return id.(int64), nil
	}
	return 0, errors.New("user_id not found in session")
}

func (s *SessionWrapper) User() (*User, error) {
	if s.user != nil {
		return s.user, nil
	}
	id, err := s.UserId()
	if err != nil {
		return nil, err
	}

	user, err := GetUserById(s.c.Request().Context(), r.DB(s.c), id)
	if err != nil {
		return nil, err
	}
	if err := LoadUserWithAccount(s.c.Request().Context(), r.DB(s.c), user); err != nil {
		return nil, err
	}

	s.user = user
	return user, nil
}

type FlashMessage struct {
	Success string
	Error   string
}

func init() {
	// internal representation of a flash message
	gob.Register(map[string]string{})
}

const flashKey = "flash"

func (s *SessionWrapper) getFlash() (map[string]string, bool) {
	if v, ok := s.s.Values[flashKey]; ok {
		return v.(map[string]string), true
	} else {
		return map[string]string{}, false
	}
}

func (s *SessionWrapper) SetFlashSuccess(msg string) error {
	flash, _ := s.getFlash()
	flash["success"] = msg
	return s.Set(flashKey, flash)
}

func (s *SessionWrapper) SetFlashError(msg string) error {
	flash, _ := s.getFlash()
	flash["error"] = msg
	return s.Set(flashKey, flash)
}

func (s *SessionWrapper) Flash() (*FlashMessage, error) {
	var err error
	flash, exists := s.getFlash()
	if exists {
		// Drop the flash data if it exists.
		delete(s.s.Values, flashKey)
		err = s.s.Save(s.c.Request(), s.c.Response())
	}

	return &FlashMessage{
		Success: flash["success"],
		Error:   flash["error"],
	}, err
}

const errorsKey = "errors"

func (s *SessionWrapper) getErrors() (map[string]string, bool) {
	if v, ok := s.s.Values[errorsKey]; ok {
		return v.(map[string]string), true
	} else {
		return map[string]string{}, false
	}
}

func (s *SessionWrapper) SetErrors(errs map[string]string) error {
	flashErrors, _ := s.getErrors()
	flashErrors = errs
	return s.Set(errorsKey, flashErrors)
}

func (s *SessionWrapper) Errors() (map[string]string, error) {
	var err error
	flashErrors, exists := s.getErrors()
	if exists {
		// Drop the flash data if it exists.
		delete(s.s.Values, errorsKey)
		err = s.s.Save(s.c.Request(), s.c.Response())
	}
	return flashErrors, err
}

func AuthRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, err := r.Session(c).User(); err != nil {
				return c.Redirect(http.StatusFound, "/login")
			}
			return next(c)
		}
	}
}

func RedirectIfAuthenticated() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, err := r.Session(c).UserId(); err == nil {
				return c.Redirect(http.StatusFound, "/")
			}
			return next(c)
		}
	}
}
