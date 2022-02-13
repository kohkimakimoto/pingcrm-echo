package app

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ContactsIndexHandler(c echo.Context) error {
	filters := &ContactFilters{}
	if err := c.Bind(filters); err != nil {
		return err
	}
	filters.AccountId = r.User(c).AccountId

	p, contacts, err := PaginateContactsByFilters(c.Request().Context(), r.DB(c), r.Pagination(c), filters)
	if err != nil {
		return err
	}

	if err := LoadContactWithOrganization(c.Request().Context(), r.DB(c), contacts...); err != nil {
		return err
	}

	return r.Inertia(c).Render(http.StatusOK, "Contacts/Index", map[string]interface{}{
		"filters": filters,
		"contacts": map[string]interface{}{
			"current_page": p.CurrentPage,
			"data": contacts.ToMap(
				"id",
				"name",
				"email",
				"phone",
				"city",
				"deleted_at",
				"organization.name",
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

func ContactsCreateHandler(c echo.Context) error {
	orgs, err := ListOrganizationsByAccountId(c.Request().Context(), r.DB(c), r.User(c).AccountId)
	if err != nil {
		return err
	}

	return r.Inertia(c).Render(http.StatusOK, "Contacts/Create", map[string]interface{}{
		"organizations": orgs.ToMap(
			"id",
			"name",
		),
	})
}

func ContactsStoreHandler(c echo.Context) error {
	input := &struct {
		FirstName      string  `json:"first_name"`
		LastName       string  `json:"last_name"`
		OrganizationId *int64  `json:"organization_id"`
		Email          *string `json:"email"`
		Phone          *string `json:"phone"`
		Address        *string `json:"address"`
		City           *string `json:"city"`
		Region         *string `json:"region"`
		Country        *string `json:"country"`
		PostalCode     *string `json:"postal_code"`
	}{}
	if err := c.Bind(input); err != nil {
		return err
	}

	// validate
	v := make(ValidationErrors)
	if input.FirstName == "" {
		v.Set("first_name", "The first name field is required.")
	}

	if input.LastName == "" {
		v.Set("last_name", "The last name field is required.")
	}

	if len(v) > 0 {
		if err := r.Session(c).SetErrors(v); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/contacts/create")
	}

	user := r.User(c)

	contact := NewContact()
	contact.AccountId = user.AccountId
	contact.OrganizationId = input.OrganizationId
	contact.FirstName = input.FirstName
	contact.LastName = input.LastName
	contact.Email = input.Email
	contact.Phone = input.Phone
	contact.Address = input.Address
	contact.City = input.City
	contact.Region = input.Region
	contact.Country = input.Country
	contact.PostalCode = input.PostalCode

	if err := CreateContact(c.Request().Context(), r.DB(c), contact); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Contact created."); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/contacts")
}

func ContactsEditHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	contact, err := GetContactById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}

	orgs, err := ListOrganizationsByAccountId(c.Request().Context(), r.DB(c), r.User(c).AccountId)
	if err != nil {
		return err
	}

	return r.Inertia(c).Render(http.StatusOK, "Contacts/Edit", map[string]interface{}{
		"contact": map[string]interface{}{
			"id":              contact.Id,
			"first_name":      contact.FirstName,
			"last_name":       contact.LastName,
			"organization_id": contact.OrganizationId,
			"email":           contact.Email,
			"phone":           contact.Phone,
			"address":         contact.Address,
			"city":            contact.City,
			"region":          contact.Region,
			"country":         contact.Country,
			"postal_code":     contact.PostalCode,
			"deleted_at":      contact.DeletedAt,
		},
		"organizations": orgs.ToMap(
			"id",
			"name",
		),
	})
}

func ContactsUpdateHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	contact, err := GetContactById(c.Request().Context(), r.DB(c), id)
	if err != nil {
		if IsErrNoRows(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			return err
		}
	}

	input := &struct {
		FirstName      string  `json:"first_name"`
		LastName       string  `json:"last_name"`
		OrganizationId *int64  `json:"organization_id"`
		Email          *string `json:"email"`
		Phone          *string `json:"phone"`
		Address        *string `json:"address"`
		City           *string `json:"city"`
		Region         *string `json:"region"`
		Country        *string `json:"country"`
		PostalCode     *string `json:"postal_code"`
	}{}
	if err := c.Bind(input); err != nil {
		return err
	}

	// validate
	v := make(ValidationErrors)
	if input.FirstName == "" {
		v.Set("first_name", "The first name field is required.")
	}

	if input.FirstName == "" {
		v.Set("last_name", "The last name field is required.")
	}

	if len(v) > 0 {
		if err := r.Session(c).SetErrors(v); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, fmt.Sprintf("/contacts/%v/edit", id))
	}

	contact.FirstName = input.FirstName
	contact.LastName = input.LastName
	contact.OrganizationId = input.OrganizationId
	contact.Email = input.Email
	contact.Phone = input.Phone
	contact.Address = input.Address
	contact.City = input.City
	contact.Region = input.Region
	contact.Country = input.Country
	contact.PostalCode = input.PostalCode

	if err := UpdateContact(c.Request().Context(), r.DB(c), contact); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Contact updated."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contacts/%v/edit", id))
}

func ContactsDestroyHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := SoftDeleteContact(c.Request().Context(), r.DB(c), id); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Contact deleted."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contacts/%v/edit", id))
}

func ContactsRestoreHandler(c echo.Context) error {
	id := ParamInt64(c, "id")
	if id == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := RestoreContact(c.Request().Context(), r.DB(c), id); err != nil {
		return err
	}

	if err := r.Session(c).SetFlashSuccess("Contact restored."); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/contacts/%v/edit", id))
}
