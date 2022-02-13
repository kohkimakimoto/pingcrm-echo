package pagination

import (
	"fmt"
	"net/url"
	"strconv"
)

type Paginator struct {
	Path                     string     // Base URL path
	Query                    url.Values // The query parameters to add to all URLs.
	PageName                 string     // The parameter name of the page.
	ItemsCount               uint64     // The number of total items.
	CurrentPage              uint64     // The current page.
	Invalid                  bool       // If the current page is not correct position, it has true.
	PerPage                  uint64     // The number of items to be shown per page.
	PagesCount               uint64     // The number of total pages
	Offset                   uint64     // The offset.
	HasPrevious              bool       // If it has previous page
	PreviousPage             uint64     // The previous page
	HasNext                  bool       // If it has next page
	NextPage                 uint64     // The next page
	NumShortcuts             uint64     // The maximum number of shortcut links.
	ShortcutPages            []uint64   // The array of the shortcut pages
	HasMorePreviousShortcuts bool       // It has more pages before shortcut links
	HasMoreNextShortcuts     bool       // It has more pages after shortcut links
}

// Url returns a URL for a given page number.
func (pg *Paginator) Url(page uint64) string {
	if page == 0 {
		page = 1
	}

	query := make(url.Values)
	for key, values := range pg.Query {
		query[key] = values
	}
	if page != 1 {
		query[pg.PageName] = []string{strconv.FormatUint(page, 10)}
	}

	urlString := pg.Path
	queryString := query.Encode()
	if queryString != "" {
		urlString += "?" + query.Encode()
	}

	return urlString
}

func (pg *Paginator) First() bool {
	return pg.CurrentPage == 1
}

func (pg *Paginator) From() uint64 {
	if pg.Invalid {
		return 0
	}

	return pg.Offset + 1
}

func (pg *Paginator) To() uint64 {
	if pg.Invalid {
		return 0
	}

	n := pg.Offset + pg.PerPage
	if n > pg.ItemsCount {
		n = pg.ItemsCount
	}
	return n
}

func (pg *Paginator) FirstPageUrl() string {
	return pg.Url(1)
}

func (pg *Paginator) LastPageUrl() string {
	return pg.Url(pg.PagesCount)
}

func (pg *Paginator) PreviousPageUrl() string {
	if !pg.HasPrevious {
		return ""
	}
	return pg.Url(pg.PreviousPage)
}

func (pg *Paginator) NextPageUrl() string {
	if !pg.HasNext {
		return ""
	}
	return pg.Url(pg.NextPage)
}

type Link struct {
	Url    string `json:"url"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

func (pg *Paginator) Links() []*Link {
	links := make([]*Link, 0)
	if pg.Invalid {
		return links
	}

	links = append(links, &Link{
		Url:    pg.PreviousPageUrl(),
		Label:  "« Previous",
		Active: false,
	})

	for _, pageN := range pg.ShortcutPages {
		links = append(links, &Link{
			Url:    pg.Url(pageN),
			Label:  strconv.FormatUint(pageN, 10),
			Active: pageN == pg.CurrentPage,
		})
	}

	links = append(links, &Link{
		Url:    pg.NextPageUrl(),
		Label:  "Next »",
		Active: false,
	})

	return links
}

func (pg *Paginator) SprintIfNotFirst(text string) string {
	if pg.CurrentPage > 1 {
		return fmt.Sprintf(text, pg.CurrentPage)
	}
	return ""
}
