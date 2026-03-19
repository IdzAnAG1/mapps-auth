-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: FindUserByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: FindUserByUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- При регистрации вызывать вместе с AssignDefaultRole в одной Go-транзакции.
-- name: CreateUser :exec
INSERT INTO users (id, email, password_hash, username)
VALUES ($1, $2, $3, $4);

-- Назначить роль buyer при регистрации.
-- name: AssignDefaultRole :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, 'role-buyer');

-- Ручное назначение произвольной роли (например, admin).
-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: GetUserRoles :many
SELECT r.id, r.name, r.description
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1;
