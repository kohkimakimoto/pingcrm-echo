package app

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/kohkimakimoto/pingcrm-echo/pkg/pagination"
)

type Contact struct {
	Id             int64         `db:"id"`
	AccountId      int64         `db:"account_id"`
	OrganizationId *int64        `db:"organization_id"`
	FirstName      string        `db:"first_name"`
	LastName       string        `db:"last_name"`
	Email          *string       `db:"email"`
	Phone          *string       `db:"phone"`
	Address        *string       `db:"address"`
	City           *string       `db:"city"`
	Region         *string       `db:"region"`
	Country        *string       `db:"country"`
	PostalCode     *string       `db:"postal_code"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
	DeletedAt      *time.Time    `db:"deleted_at"`
	Organization   *Organization `db:"-"`
}

func NewContact() *Contact {
	return &Contact{}
}

func (c *Contact) ToMap(keys ...string) map[string]interface{} {
	m := make(map[string]interface{})
	for _, k := range keys {
		switch k {
		case "id":
			m["id"] = c.Id
		case "account_id":
			m["account_id"] = c.AccountId
		case "organization_id":
			m["organization_id"] = c.OrganizationId
		case "first_name":
			m["first_name"] = c.FirstName
		case "last_name":
			m["last_name"] = c.LastName
		case "email":
			m["email"] = c.Email
		case "phone":
			m["phone"] = c.Phone
		case "address":
			m["address"] = c.Address
		case "city":
			m["city"] = c.City
		case "region":
			m["region"] = c.Region
		case "country":
			m["country"] = c.Country
		case "postal_code":
			m["postal_code"] = c.PostalCode
		case "created_at":
			m["created_at"] = c.CreatedAt
		case "updated_at":
			m["updated_at"] = c.UpdatedAt
		case "deleted_at":
			m["deleted_at"] = c.DeletedAt
		case "organization.name":
			if c.Organization != nil {
				if m["organization"] == nil {
					m["organization"] = make(map[string]interface{})
				}
				m["organization"].(map[string]interface{})["name"] = c.Organization.Name
			}
		case "name":
			m["name"] = c.Name()
		default:
			panic("unknown key: " + k)
		}
	}
	return m
}

func (c *Contact) Name() string {
	return c.FirstName + " " + c.LastName
}

type ContactList []*Contact

func (l ContactList) ToMap(keys ...string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(l))
	for i, v := range l {
		result[i] = v.ToMap(keys...)
	}
	return result
}

func (l ContactList) Ids() []int64 {
	ids := make([]int64, len(l))
	for i, v := range l {
		ids[i] = v.Id
	}
	return ids
}

func (l ContactList) OrganizationIds() []int64 {
	ids := make([]int64, 0)
	for _, v := range l {
		if v.OrganizationId != nil {
			ids = append(ids, *v.OrganizationId)
		}
	}
	return ids
}

func (l ContactList) ByIds() map[int64]*Contact {
	result := make(map[int64]*Contact, len(l))
	for _, v := range l {
		result[v.Id] = v
	}
	return result
}

func (l ContactList) ByOrganizationIds() map[int64]ContactList {
	result := make(map[int64]ContactList)
	for _, v := range l {
		if v.OrganizationId != nil {
			if _, ok := result[*v.OrganizationId]; !ok {
				result[*v.OrganizationId] = ContactList{}
			}
			result[*v.OrganizationId] = append(result[*v.OrganizationId], v)
		}
	}
	return result
}

func GetContactByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (*Contact, error) {
	dest := NewContact()
	if err := db.GetContextByBuilder(ctx, dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func ListContactsByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (ContactList, error) {
	dest := make(ContactList, 0)
	if err := db.SelectContextByBuilder(ctx, &dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func LoadContactWithOrganization(ctx context.Context, db *DB, contacts ...*Contact) error {
	q := sq.Select("*").From("organizations").Where(
		sq.And{
			sq.Expr("deleted_at IS NULL"),
			sq.Eq{"id": ContactList(contacts).OrganizationIds()},
		},
	)

	orgs, err := ListOrganizationsByBuilder(ctx, db, q)
	if err != nil {
		return err
	}
	byIds := orgs.ByIds()

	for _, con := range contacts {
		if con.OrganizationId != nil {
			if org, ok := byIds[*con.OrganizationId]; ok {
				con.Organization = org
			}
		}
	}
	return nil
}

type ContactFilters struct {
	Search    string `json:"search" query:"search"`
	Trashed   string `json:"trashed" query:"trashed"`
	AccountId int64  `json:"-" query:"-"`
}

func PaginateContactsByFilters(ctx context.Context, db *DB, p *pagination.Factory, filter *ContactFilters) (*pagination.Paginator, ContactList, error) {
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
		cond = append(cond, sq.Or{
			sq.Like{"first_name": "%" + filter.Search + "%"},
			sq.Like{"last_name": "%" + filter.Search + "%"},
			sq.Like{"email": "%" + filter.Search + "%"},
			sq.Expr(`exists (select * from organizations where contacts.organization_id = organizations.id and name like ? and organizations.deleted_at is null)`, "%"+filter.Search+"%"),
		})
	}

	count, err := db.CountContextByBuilder(ctx, sq.Select("COUNT(*) AS count").
		From("contacts").
		Where(sq.And(cond)))
	if err != nil {
		return nil, nil, err
	}

	pg := p.Paginator(count)
	if pg.Invalid {
		return pg, make(ContactList, 0), nil
	}

	q := sq.Select("*").From("contacts").
		Where(sq.And(cond)).
		OrderBy("contacts.last_name ASC, contacts.first_name ASC").
		Offset(pg.Offset).
		Limit(pg.PerPage)
	dest, err := ListContactsByBuilder(ctx, db, q)
	if err != nil {
		return nil, nil, err
	}
	return pg, dest, nil
}

func GetContactById(ctx context.Context, db *DB, id int64) (*Contact, error) {
	q := sq.Select("*").From("contacts").Where(sq.Eq{"id": id})
	return GetContactByBuilder(ctx, db, q)
}

func CreateContact(ctx context.Context, db *DB, contact *Contact) error {
	contact.CreatedAt = time.Now()
	contact.UpdatedAt = time.Now()
	q := sq.Insert("contacts").
		Columns(
			"account_id",
			"organization_id",
			"first_name",
			"last_name",
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
			contact.AccountId,
			contact.OrganizationId,
			contact.FirstName,
			contact.LastName,
			contact.Email,
			contact.Phone,
			contact.Address,
			contact.City,
			contact.Region,
			contact.Country,
			contact.PostalCode,
			contact.CreatedAt,
			contact.UpdatedAt,
		)
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func UpdateContact(ctx context.Context, db *DB, contact *Contact) error {
	contact.UpdatedAt = time.Now()
	q := sq.Update("contacts").
		Set("account_id", contact.AccountId).
		Set("organization_id", contact.OrganizationId).
		Set("first_name", contact.FirstName).
		Set("last_name", contact.LastName).
		Set("email", contact.Email).
		Set("phone", contact.Phone).
		Set("address", contact.Address).
		Set("city", contact.City).
		Set("region", contact.Region).
		Set("country", contact.Country).
		Set("postal_code", contact.PostalCode).
		Set("updated_at", contact.UpdatedAt).
		Where(sq.Eq{"id": contact.Id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func SoftDeleteContact(ctx context.Context, db *DB, id int64) error {
	q := sq.Update("contacts").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func RestoreContact(ctx context.Context, db *DB, id int64) error {
	q := sq.Update("contacts").
		Set("deleted_at", nil).
		Where(sq.Eq{"id": id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}
