-- name: ListAccounts :many
SELECT *
FROM accounts
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(status)::account_status_enum IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(name_query)::text IS NULL OR name ILIKE ('%' || sqlc.narg(name_query) || '%'))
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CountAccounts :one
SELECT count(*)::bigint
FROM accounts
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(status)::account_status_enum IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(name_query)::text IS NULL OR name ILIKE ('%' || sqlc.narg(name_query) || '%'));

-- name: CreateAccount :one
INSERT INTO accounts (
  tenant_id,
  owner_user_id,
  name,
  industry,
  website,
  phone,
  status,
  memo,
  created_by
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(owner_user_id),
  sqlc.arg(name),
  sqlc.narg(industry),
  sqlc.narg(website),
  sqlc.narg(phone),
  coalesce(sqlc.narg(status)::account_status_enum, 'prospect'),
  sqlc.narg(memo),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts
SET
  owner_user_id = coalesce(sqlc.narg(owner_user_id), owner_user_id),
  name = coalesce(sqlc.narg(name), name),
  industry = coalesce(sqlc.narg(industry), industry),
  website = coalesce(sqlc.narg(website), website),
  phone = coalesce(sqlc.narg(phone), phone),
  status = coalesce(sqlc.narg(status)::account_status_enum, status),
  memo = coalesce(sqlc.narg(memo), memo),
  updated_at = now()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(account_id)
RETURNING *;

-- name: ListContactsByAccount :many
SELECT *
FROM contacts
WHERE tenant_id = sqlc.arg(tenant_id)
  AND account_id = sqlc.arg(account_id)
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CreateContact :one
INSERT INTO contacts (
  tenant_id,
  account_id,
  location_id,
  owner_user_id,
  full_name,
  department,
  title,
  email,
  phone,
  is_primary,
  memo,
  created_by
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(account_id),
  sqlc.narg(location_id),
  sqlc.arg(owner_user_id),
  sqlc.arg(full_name),
  sqlc.narg(department),
  sqlc.narg(title),
  sqlc.narg(email),
  sqlc.narg(phone),
  coalesce(sqlc.narg(is_primary), false),
  sqlc.narg(memo),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: ListLocationsByAccount :many
SELECT *
FROM account_locations
WHERE tenant_id = sqlc.arg(tenant_id)
  AND account_id = sqlc.arg(account_id)
ORDER BY updated_at DESC;

-- name: CreateLocation :one
INSERT INTO account_locations (
  tenant_id,
  account_id,
  name,
  country,
  postal_code,
  prefecture,
  city,
  address_line1,
  address_line2
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(account_id),
  sqlc.arg(name),
  sqlc.narg(country),
  sqlc.narg(postal_code),
  sqlc.narg(prefecture),
  sqlc.narg(city),
  sqlc.narg(address_line1),
  sqlc.narg(address_line2)
)
RETURNING *;
