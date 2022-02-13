CREATE TABLE IF NOT EXISTS "personal_access_tokens" (
	"id"	integer NOT NULL,
	"tokenable_type"	varchar NOT NULL,
	"tokenable_id"	integer NOT NULL,
	"name"	varchar NOT NULL,
	"token"	varchar NOT NULL,
	"abilities"	text,
	"last_used_at"	datetime,
	"created_at"	datetime,
	"updated_at"	datetime,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE TABLE IF NOT EXISTS "password_resets" (
	"email"	varchar NOT NULL,
	"token"	varchar NOT NULL,
	"created_at"	datetime
);
CREATE TABLE IF NOT EXISTS "failed_jobs" (
	"id"	integer NOT NULL,
	"uuid"	varchar NOT NULL,
	"connection"	text NOT NULL,
	"queue"	text NOT NULL,
	"payload"	text NOT NULL,
	"exception"	text NOT NULL,
	"failed_at"	datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE TABLE IF NOT EXISTS "accounts" (
	"id"	integer NOT NULL,
	"name"	varchar NOT NULL,
	"created_at"	datetime,
	"updated_at"	datetime,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE TABLE IF NOT EXISTS "users" (
	"id"	integer NOT NULL,
	"account_id"	integer NOT NULL,
	"first_name"	varchar NOT NULL,
	"last_name"	varchar NOT NULL,
	"email"	varchar NOT NULL,
	"email_verified_at"	datetime,
	"password"	varchar,
	"owner"	tinyint(1) NOT NULL DEFAULT '0',
	"photo_path"	varchar,
	"remember_token"	varchar,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE TABLE IF NOT EXISTS "organizations" (
	"id"	integer NOT NULL,
	"account_id"	integer NOT NULL,
	"name"	varchar NOT NULL,
	"email"	varchar,
	"phone"	varchar,
	"address"	varchar,
	"city"	varchar,
	"region"	varchar,
	"country"	varchar,
	"postal_code"	varchar,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE TABLE IF NOT EXISTS "contacts" (
	"id"	integer NOT NULL,
	"account_id"	integer NOT NULL,
	"organization_id"	integer,
	"first_name"	varchar NOT NULL,
	"last_name"	varchar NOT NULL,
	"email"	varchar,
	"phone"	varchar,
	"address"	varchar,
	"city"	varchar,
	"region"	varchar,
	"country"	varchar,
	"postal_code"	varchar,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE INDEX IF NOT EXISTS "personal_access_tokens_tokenable_type_tokenable_id_index" ON "personal_access_tokens" (
	"tokenable_type",
	"tokenable_id"
);
CREATE UNIQUE INDEX IF NOT EXISTS "personal_access_tokens_token_unique" ON "personal_access_tokens" (
	"token"
);
CREATE INDEX IF NOT EXISTS "password_resets_email_index" ON "password_resets" (
	"email"
);
CREATE UNIQUE INDEX IF NOT EXISTS "failed_jobs_uuid_unique" ON "failed_jobs" (
	"uuid"
);
CREATE INDEX IF NOT EXISTS "users_account_id_index" ON "users" (
	"account_id"
);
CREATE UNIQUE INDEX IF NOT EXISTS "users_email_unique" ON "users" (
	"email"
);
CREATE INDEX IF NOT EXISTS "organizations_account_id_index" ON "organizations" (
	"account_id"
);
CREATE INDEX IF NOT EXISTS "contacts_account_id_index" ON "contacts" (
	"account_id"
);
CREATE INDEX IF NOT EXISTS "contacts_organization_id_index" ON "contacts" (
	"organization_id"
);
