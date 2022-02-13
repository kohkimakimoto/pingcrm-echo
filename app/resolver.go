package app

import (
	"github.com/kohkimakimoto/inertia-echo"
	"github.com/labstack/echo/v4"

	"github.com/kohkimakimoto/pingcrm-echo/pkg/pagination"
)

var r = &resolver{}

// resolver is a namespace to provide helper functions that resolve objects
type resolver struct{}

func (r *resolver) DB(c echo.Context) *DB {
	return getApp(c).DB
}

func (r *resolver) Inertia(c echo.Context) *inertia.Inertia {
	return inertia.MustGet(c)
}

func (r *resolver) Session(c echo.Context) *SessionWrapper {
	s, err := NewSessionWrapper(c)
	if err != nil {
		panic(err)
	}
	return s
}

func (r *resolver) User(c echo.Context) *User {
	user, err := r.Session(c).User()
	if err != nil {
		panic(err)
	}
	return user
}

func (r *resolver) Pagination(c echo.Context) *pagination.Factory {
	return pagination.NewFactory(c)
}

func (r *resolver) IsDemo(c echo.Context) bool {
	return getApp(c).Config.IsDemo
}
