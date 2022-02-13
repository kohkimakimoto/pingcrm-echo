package app

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func OrganizationsIndexHandler(c echo.Context) error {
	filters := &OrganizationFilters{}
	if err := c.Bind(filters); err != nil {
		return err
	}
	filters.AccountId = r.User(c).AccountId

	p, orgs, err := PaginateOrganizationsFilters(c.Request().Context(), r.DB(c), r.Pagination(c), filters)
	if err != nil {
		return err
	}

	return r.Inertia(c).Render(http.StatusOK, "Organizations/Index", map[string]interface{}{
		"filters": filters,
		"organizations": map[string]interface{}{
			"current_page": p.CurrentPage,
			"data": orgs.ToMap(
				"id",
				"name",
				"phone",
				"city",
				"deleted_at",
			),
			"links":          p.Links(),
			"path":           p.Path,
			"per_page":       p.PerPage,
			"prev_page_url":  p.PreviousPageUrl(),
			"next_page_url":  p.NextPageUrl(),
			"first_page_url": p.FirstPageUrl(),
			"last_page_url":  p.LastPageUrl(),
			"last_page":      p.PagesCount,
			"from":           p.From(),
			"to":             p.To(),
			"total":          p.ItemsCount,
		},
	})
}

func OrganizationsCreateHandler(c echo.Context) error {
	return r.Inertia(c).Render(http.StatusOK, "Organizations/Create", nil)
}

func OrganizationsStoreHandler(c echo.Context) error {
	input := &struct {
		Name       string  `json:"name"`
		Email      *string `json:"email"`
		Phone      *string `json:"phone"`
		Address    *string `json:"address"`
		City       *string `json:"city"`
		Region     *string `json:"region"`
		Country    *string `json:"country"`
		PostalCode *string `json:"postal_code"`
	}{}
	if err := c.Bind(input); err != nil {
		return err
	}

	// validate
	v := make(ValidationErrors)
	if input.Name == "" {
		v.Set("name", "The name field is required.")
	}

	if len(v) > 0 {
		if err := r.Session(c).SetErrors(v); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/organizations/create")
	}

	user := r.User(c)

	org := NewOrganization()
	org.AccountId = user.AccountId
	org.Name = input.Name
	org.Email = input.Email
	org.Phone = input.Phone
	org.Address = input.Address
	org.City = input.City
	org.Region = input.Region
	org.Country = input.Country
	org.PostalCode = input.PostalCode

	if err := CreateOrganization(c.Request().Context(), r.DB(c), org); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Organization created."); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/organizations")
}

func OrganizationsEditHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	org, err := GetOrganizationById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}

	if err := LoadOrganizationWithContacts(c.Request().Context(), r.DB(c), org); err != nil {
		return err
	}

	return r.Inertia(c).Render(http.StatusOK, "Organizations/Edit", map[string]interface{}{
		"organization": map[string]interface{}{
			"id":          org.Id,
			"name":        org.Name,
			"email":       org.Email,
			"phone":       org.Phone,
			"address":     org.Address,
			"city":        org.City,
			"region":      org.Region,
			"country":     org.Country,
			"postal_code": org.PostalCode,
			"deleted_at":  org.DeletedAt,
			"contacts": org.Contacts.ToMap(
				"id",
				"name",
				"city",
				"phone",
			),
		},
	})
}

func OrganizationsUpdateHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	org, err := GetOrganizationById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}

	input := &struct {
		Name       string  `json:"name"`
		Email      *string `json:"email"`
		Phone      *string `json:"phone"`
		Address    *string `json:"address"`
		City       *string `json:"city"`
		Region     *string `json:"region"`
		Country    *string `json:"country"`
		PostalCode *string `json:"postal_code"`
	}{}
	if err := c.Bind(input); err != nil {
		return err
	}

	// validate
	v := make(ValidationErrors)
	if input.Name == "" {
		v.Set("name", "The name field is required.")
	}

	if len(v) > 0 {
		if err := r.Session(c).SetErrors(v); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, fmt.Sprintf("/organizations/%v/edit", id))
	}

	org.Name = input.Name
	org.Email = input.Email
	org.Phone = input.Phone
	org.Address = input.Address
	org.City = input.City
	org.Region = input.Region
	org.Country = input.Country
	org.PostalCode = input.PostalCode

	if err := UpdateOrganization(c.Request().Context(), r.DB(c), org); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Organization updated."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/organizations/%v/edit", id))
}

func OrganizationsDestroyHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := SoftDeleteOrganization(c.Request().Context(), r.DB(c), id); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Organization deleted."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/organizations/%v/edit", id))
}

func OrganizationsRestoreHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := RestoreOrganization(c.Request().Context(), r.DB(c), id); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Organization restored."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/organizations/%v/edit", id))
}
