package pagination

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Factory is an object that can create a Paginator.
type Factory struct {
	c         echo.Context // c is an echo.Context to get current page from query string or URL.
	PageName  string       // PageName is a query string name to get current page from query string.
	PerPage   uint64       // PerPage is the number of items per page.
	Shortcuts uint64       // Shortcuts must be odd number
}

func NewFactory(c echo.Context) *Factory {
	return &Factory{
		c:         c,
		PageName:  "page",
		PerPage:   10,
		Shortcuts: 10,
	}
}

type InvalidPageError struct {
	CurrentPage uint64
	PagesCount  uint64
}

func (e *InvalidPageError) Error() string {
	return fmt.Sprintf("the current page '%d', but the number of total pages is '%d'.", e.CurrentPage, e.PagesCount)
}

func IsInvalidPageError(err error) bool {
	_, ok := err.(*InvalidPageError)
	return ok
}

func (f *Factory) Paginator(count uint64) *Paginator {
	c := f.c
	req := c.Request()

	pg := &Paginator{}
	pg.Path = req.URL.Path
	pg.PageName = f.PageName
	if f.Shortcuts%2 != 1 {
		pg.NumShortcuts = f.Shortcuts + 1
	} else {
		pg.NumShortcuts = f.Shortcuts
	}

	pg.Query = make(url.Values)
	for key, values := range req.URL.Query() {
		if pg.PageName != key {
			// keeps query strings except PageName key
			pg.Query[key] = values
		}
	}

	pg.CurrentPage = getCurrentPage(c, pg.PageName, 1)
	pg.PerPage = f.PerPage
	pg.ItemsCount = count

	// calculate pages count and offset
	if pg.ItemsCount == 0 {
		pg.PagesCount = 0
		pg.Offset = 0
	} else {
		pg.PagesCount = ((pg.ItemsCount - 1) / pg.PerPage) + 1
		pg.Offset = (pg.CurrentPage - 1) * pg.PerPage
	}

	// validate the current page
	if pg.CurrentPage == 0 || pg.CurrentPage > pg.PagesCount {
		// The current page is invalid.
		pg.Invalid = true
		return pg
	}

	// check the previous page existence.
	if pg.CurrentPage > 1 {
		pg.HasPrevious = true
		pg.PreviousPage = pg.CurrentPage - 1
	} else {
		pg.HasPrevious = false
		pg.PreviousPage = 0
	}

	// check the nest page existence.
	if pg.CurrentPage < pg.PagesCount {
		pg.HasNext = true
		pg.NextPage = pg.CurrentPage + 1
	} else {
		pg.HasNext = false
	}

	// construct Shortcuts.
	// The Shortcuts are links to a specific page.
	quotient := pg.NumShortcuts / 2
	if pg.PagesCount <= pg.NumShortcuts {
		// In this situation, the paginator should make the shortcut links to all pages.
		for i := uint64(1); i <= pg.PagesCount; i++ {
			pg.ShortcutPages = append(pg.ShortcutPages, i)
		}
		pg.HasMorePreviousShortcuts = false
		pg.HasMoreNextShortcuts = false
	} else if pg.CurrentPage <= (1 + quotient) {
		// The current page is in early position.
		for i := uint64(1); i <= pg.NumShortcuts; i++ {
			pg.ShortcutPages = append(pg.ShortcutPages, i)
		}
		pg.HasMorePreviousShortcuts = false
		pg.HasMoreNextShortcuts = true
	} else if pg.CurrentPage >= (pg.PagesCount - quotient) {
		// The current page is in late position.
		for i := pg.PagesCount - pg.NumShortcuts + 1; i <= pg.PagesCount; i++ {
			pg.ShortcutPages = append(pg.ShortcutPages, i)
		}
		pg.HasMorePreviousShortcuts = true
		pg.HasMoreNextShortcuts = false
	} else {
		// The current page is in middle position.
		for i := pg.CurrentPage - quotient; i <= pg.CurrentPage+quotient; i++ {
			pg.ShortcutPages = append(pg.ShortcutPages, i)
		}
		pg.HasMorePreviousShortcuts = true
		pg.HasMoreNextShortcuts = true
	}

	return pg
}

func getCurrentPage(c echo.Context, name string, defaultPage ...uint64) uint64 {
	if v := getParam(c, name); len(v) > 0 {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}

	if len(defaultPage) > 0 {
		return defaultPage[0]
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
