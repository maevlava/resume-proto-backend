-- name: CreateUser :one
INSERT INTO users(id, name, email, hashed_password)
VALUES(
       $1,
       $2,
       $3,
       $4
      )
RETURNING *;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE users.id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE users.email = $1;

-- name: UpdateUser :one
UPDATE users
SET
    updated_at = now(),
    name = $2,
    email = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE users.id = $1;