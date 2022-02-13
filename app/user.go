package app

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	sq "github.com/Masterminds/squirrel"
)

type User struct {
	Id              int64      `db:"id"`
	AccountId       int64      `db:"account_id"`
	FirstName       string     `db:"first_name"`
	LastName        string     `db:"last_name"`
	Email           string     `db:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	Password        string     `db:"password"`
	Owner           bool       `db:"owner"`
	PhotoPath       *string    `db:"photo_path"`
	RememberToken   string     `db:"remember_token"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	Account         *Account   `db:"-"`
}

func NewUser() *User {
	return &User{}
}

func (u *User) ToMap(keys ...string) map[string]interface{} {
	m := make(map[string]interface{})
	for _, k := range keys {
		switch k {
		case "id":
			m["id"] = u.Id
		case "account_id":
			m["account_id"] = u.AccountId
		case "first_name":
			m["first_name"] = u.FirstName
		case "last_name":
			m["last_name"] = u.LastName
		case "email":
			m["email"] = u.Email
		case "email_verified_at":
			m["email_verified_at"] = u.EmailVerifiedAt
		case "password":
			m["password"] = u.Password
		case "owner":
			m["owner"] = u.Owner
		case "photo_path":
			m["photo_path"] = u.PhotoPath
		case "remember_token":
			m["remember_token"] = u.RememberToken
		case "created_at":
			m["created_at"] = u.CreatedAt
		case "updated_at":
			m["updated_at"] = u.UpdatedAt
		case "deleted_at":
			m["deleted_at"] = u.DeletedAt
		case "name":
			m["name"] = u.Name()
		case "photo":
			// not implemented yet.
			m["photo"] = nil
		default:
			panic("unknown key: " + k)
		}
	}
	return m
}

func (u *User) IsDemoUser() bool {
	return u.Email == "johndoe@example.com"
}

func (u *User) Name() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) VerifyPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain))
	return err == nil
}

func (u *User) SetPlainPassword(plain string) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	u.Password = string(h)
}

type UserList []*User

func (l UserList) ToMap(keys ...string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(l))
	for i, v := range l {
		result[i] = v.ToMap(keys...)
	}
	return result
}

func (l UserList) Ids() []int64 {
	ids := make([]int64, len(l))
	for i, v := range l {
		ids[i] = v.Id
	}
	return ids
}

func GetUserByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (*User, error) {
	dest := &User{}
	if err := db.GetContextByBuilder(ctx, dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func ListUsersByBuilder(ctx context.Context, db *DB, builder QueryBuilder) (UserList, error) {
	dest := make(UserList, 0)
	if err := db.SelectContextByBuilder(ctx, &dest, builder); err != nil {
		return nil, err
	}
	return dest, nil
}

func GetUserByEmail(ctx context.Context, db *DB, email string) (*User, error) {
	q := sq.Select("*").From("users").Where(sq.Eq{"email": email})
	return GetUserByBuilder(ctx, db, q)
}

func GetUserById(ctx context.Context, db *DB, id int64) (*User, error) {
	q := sq.Select("*").From("users").Where(sq.Eq{"id": id})
	return GetUserByBuilder(ctx, db, q)
}

func LoadUserWithAccount(ctx context.Context, db *DB, user *User) error {
	if user.Account != nil {
		return nil
	}
	if user.AccountId == 0 {
		return nil
	}

	account, err := GetAccountById(ctx, db, user.AccountId)
	if err != nil {
		return err
	}
	user.Account = account
	return nil
}

type UserFilters struct {
	Search    string `json:"search" query:"search"`
	Role      string `json:"role" query:"role"`
	Trashed   string `json:"trashed" query:"trashed"`
	AccountId int64  `json:"-" query:"-"`
}

func ListUsersByFilters(ctx context.Context, db *DB, filter *UserFilters) (UserList, error) {
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
		})
	}

	if filter.Role == "user" {
		cond = append(cond, sq.Eq{"owner": false})
	} else if filter.Role == "owner" {
		cond = append(cond, sq.Eq{"owner": true})
	}

	q := sq.Select("*").From("users").
		Where(sq.And(cond)).
		OrderBy("last_name ASC, first_name ASC")
	return ListUsersByBuilder(ctx, db, q)
}

func CreateUser(ctx context.Context, db *DB, user *User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	q := sq.Insert("users").
		Columns(
			"account_id",
			"first_name",
			"last_name",
			"email",
			"email_verified_at",
			"password",
			"owner",
			"photo_path",
			"remember_token",
			"created_at",
			"updated_at").
		Values(
			user.AccountId,
			user.FirstName,
			user.LastName,
			user.Email,
			user.EmailVerifiedAt,
			user.Password,
			user.Owner,
			user.PhotoPath,
			user.RememberToken,
			user.CreatedAt,
			user.UpdatedAt,
		)
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func UpdateUser(ctx context.Context, db *DB, user *User) error {
	user.UpdatedAt = time.Now()
	q := sq.Update("users").
		Set("account_id", user.AccountId).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("email", user.Email).
		Set("email_verified_at", user.EmailVerifiedAt).
		Set("password", user.Password).
		Set("owner", user.Owner).
		Set("photo_path", user.PhotoPath).
		Set("remember_token", user.RememberToken).
		Set("updated_at", user.UpdatedAt).
		Where(sq.Eq{"id": user.Id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func SoftDeleteUser(ctx context.Context, db *DB, id int64) error {
	q := sq.Update("users").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}

func RestoreUser(ctx context.Context, db *DB, id int64) error {
	q := sq.Update("users").
		Set("deleted_at", nil).
		Where(sq.Eq{"id": id})
	if _, err := db.ExecContextByBuilder(ctx, q); err != nil {
		return err
	}
	return nil
}
