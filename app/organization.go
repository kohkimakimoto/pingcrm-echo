package app

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/kohkimakimoto/pingcrm-echo/pkg/pagination"
)

type Organization struct {
	Id         int64       `db:"id"`
	AccountId  int64       `db:"account_id"`
	Name       string      `db:"name"`
	Email      *string     `db:"email"`
	Phone      *string     `db:"phone"`
	Address    *string     `db:"address"`
	City       *string     `db:"city"`
	Region     *string     `db:"region"`
	Country    *string     `db:"country"`
	PostalCode *string     `db:"postal_code"`
	CreatedAt  time.Time   `db:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at"`
	DeletedAt  *time.Time  `db:"deleted_at"`
	Contacts   ContactList `db:"-"`
}

func NewOrganization() *Organization {
	return &Organization{
		Contacts: make([]*Contact, 0),
	}
}

func (o *Organization) ToMap(keys ...string) map[string]interface{} {
	m := make(map[string]interface{})
	for _, k := range keys {
		switch k {
		case "id":
			m["id"] = o.Id
		case "account_id":
			m["account_id"] = o.AccountId
		case "name":
			m["name"] = o.Name
		case "email":
			m["email"] = o.Email
		case "phone":
			m["phone"] = o.Phone
		case "address":
			m["address"] = o.Address
		case "city":
			m["city"] = o.City
		case "region":
			m["region"] = o.Region
		case "country":
			m["country"] = o.Country
		case "postal_code":
			m["postal_code"] = o.PostalCode
		case "created_at":
			m["created_at"] = o.CreatedAt
		case "updated_at":
			m["updated_at"] = o.UpdatedAt
		case "deleted_at":
			m["deleted_at"] = o.DeletedAt
		case "contacts":
			m["contacts"] = o.Contacts
		default:
			panic("unknown key: " + k)
		}
	}
	return m
}

type OrganizationList []*Organization

func (l OrganizationList) ToMap(keys ...string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(l))
	for i, v := range l {
		result[i] = v.ToMap(keys...)
	}
	return result
}

func (l OrganizationList) Ids() []int64 {
	ids := make([]int64, len(l))
	for i, v := range l {
		ids[i] = v.Id
	}
	return ids
}

func (l OrganizationList) ByIds() map[int64]*Organization {
	result := make(map[int64]*Organization, len(l))
	for _, v := range l {
		result[v.Id] = v
	}
	return result
}

func GetOrganizationByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (*Organization, error) {
	dest := NewOrganization()
	if err := db.GetContextByBuilder(ctx, dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func ListOrganizationsByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (OrganizationList, error) {
	dest := make(OrganizationList, 0)
	if err := db.SelectContextByBuilder(ctx, &dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func LoadOrganizationWithContacts(ctx context.Context, db *DB, orgs ...*Organization) error {
	q := sq.Select("*").From("contacts").Where(
		sq.And{
			sq.Expr("deleted_at IS NULL"),
			sq.Eq{"organization_id": OrganizationList(orgs).Ids()},
		},
	).OrderBy("last_name asc, first_name asc")

	contacts, err := ListContactsByBuilder(ctx, db, q)
	if err != nil {
		return err
	}
	byOrganizationIds := contacts.ByOrganizationIds()

	for _, org := range orgs {
		if contacts, ok := byOrganizationIds[org.Id]; ok {
			org.Contacts = contacts
		} else {
			org.Contacts = make(ContactList, 0)
		}
	}
	return nil
}

type OrganizationFilters struct {
	Search    string `json:"search" query:"search"`
	Trashed   string `json:"trashed" query:"trashed"`
	AccountId int64  `json:"-" query:"-"`
}

func PaginateOrganizationsFilters(ctx context.Context, db *DB, p *pagination.Factory, filter *OrganizationFilters) (*pagination.Paginator, OrganizationList, error) {
	var cond []sq.Sqlizer

	if filter.AccountId != 0 {
		cond = append(cond, sq.Eq{"account_id": filter.AccountId})
	}

	if filter.Trashed == "only" {
		cond = append(cond, sq.Expr("deleted_at IS NOT NULL"))
	} else if filter.Trashed == "with" {
		cond = append(cond, sq.Expr("1 = 1"))
	} else {
		cond = append(cond, sq.Expr("deleted_at IS NULL"))
	}

	if filter.Search != "" {
		cond = append(cond, sq.Like{"name": "%" + filter.Search + "%"})
	}

	count, err := db.CountContextByBuilder(ctx, sq.Select("COUNT(*) AS count").From("organizations").Where(sq.And(cond)))
	if err != nil {
		return nil, nil, err
	}

	pg := p.Paginator(count)
	if pg.Invalid {
		return pg, make(OrganizationList, 0), nil
	}

	q := sq.Select("*").From("organizations").Where(sq.And(cond)).
		OrderBy("name ASC").
		Offset(pg.Offset).
		Limit(pg.PerPage)
	dest, err := ListOrganizationsByBuilder(ctx, db, q)
	if err != nil {
		return nil, nil, err
	}
	return pg, dest, nil
}

func ListOrganizationsByAccountId(ctx context.Context, db *DB, accountId int64) (OrganizationList, error) {
	q := sq.Select("*").From("organizations").Where(
		sq.And{
			sq.Expr("deleted_at IS NULL"),
			sq.Eq{"account_id": accountId},
		},
	).OrderBy("name ASC")
	return ListOrganizationsByBuilder(ctx, db, q)
}

func GetOrganizationById(ctx context.Context, db *DB, id int64) (*Organization, error) {
	q := sq.Select("*").From("organizations").Where(sq.Eq{"id": id})
	return GetOrganizationByBuilder(ctx, db, q)
}

func CreateOrganization(ctx context.Context, db *DB, org *Organization) error {
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()
	q := sq.Insert("organizations").
		Columns(
			"account_id",
			"name",
			"email",
			"phone",
			"address",
			"city",
			"region",
			"country",
			"postal_code",
			"created_at",
			"updated_at",
		).
		Values(
			org.AccountId,
			org.Name,
			org.Email,
			org.Phone,
			org.Address,
			org.City,
			org.Region,
			org.Country,
			org.PostalCode,
			org.CreatedAt,
			org.UpdatedAt,
		)
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func UpdateOrganization(ctx context.Context, db *DB, org *Organization) error {
	org.UpdatedAt = time.Now()
	q := sq.Update("organizations").
		Set("account_id", org.AccountId).
		Set("name", org.Name).
		Set("email", org.Email).
		Set("phone", org.Phone).
		Set("address", org.Address).
		Set("city", org.City).
		Set("region", org.Region).
		Set("country", org.Country).
		Set("postal_code", org.PostalCode).
		Set("updated_at", org.UpdatedAt).
		Where(sq.Eq{"id": org.Id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func SoftDeleteOrganization(ctx context.Context, db *DB, id int64) error {
	q := sq.Update("organizations").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func RestoreOrganization(ctx context.Context, db *DB, id int64) error {
	q := sq.Update("organizations").
		Set("deleted_at", nil).
		Where(sq.Eq{"id": id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}
