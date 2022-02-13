package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestFactory_Paginator(t *testing.T) {
	t.Run("Paginator with page = 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)

		factory := NewFactory(c)
		factory.PerPage = 30
		factory.Shortcuts = 5
		pg := factory.Paginator(100)
		assert.Equal(t, uint64(30), pg.PerPage)
		assert.Equal(t, uint64(4), pg.PagesCount)
		assert.Equal(t, uint64(1), pg.CurrentPage)
		assert.Equal(t, uint64(0), pg.Offset)
		assert.Equal(t, "/", pg.Path)
	})

	t.Run("Paginator with page = 2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/foo/bar?page=2", nil)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)

		factory := NewFactory(c)
		factory.PerPage = 30
		factory.Shortcuts = 5
		pg := factory.Paginator(100)
		assert.Equal(t, uint64(30), pg.PerPage)
		assert.Equal(t, uint64(4), pg.PagesCount)
		assert.Equal(t, uint64(2), pg.CurrentPage)
		assert.Equal(t, uint64(30), pg.Offset)
		assert.Equal(t, "/foo/bar", pg.Path)
	})

	t.Run("Paginator with custom PageName and PerPage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/foo/bar?pg=2", nil)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)

		factory := NewFactory(c)
		factory.PageName = "pg"
		factory.PerPage = 10
		factory.Shortcuts = 5
		pg := factory.Paginator(100)
		assert.Equal(t, uint64(10), pg.PerPage)
		assert.Equal(t, uint64(10), pg.PagesCount)
		assert.Equal(t, uint64(2), pg.CurrentPage)
		assert.Equal(t, uint64(10), pg.Offset)
		assert.Equal(t, "/foo/bar", pg.Path)
	})

	t.Run("Paginator with query strings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/foo/bar?foo=v1&foo=v2&bar=v2&page=2", nil)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)

		factory := NewFactory(c)
		factory.PerPage = 30
		factory.Shortcuts = 5
		pg := factory.Paginator(100)
		assert.Equal(t, uint64(30), pg.PerPage)
		assert.Equal(t, uint64(4), pg.PagesCount)
		assert.Equal(t, uint64(2), pg.CurrentPage)
		assert.Equal(t, uint64(30), pg.Offset)
		assert.Equal(t, "/foo/bar", pg.Path)
		assert.Regexp(t, `\/foo\/bar\?.*page\=3`, pg.Url(3))
	})
}
