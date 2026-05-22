-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, bio)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetUserById :one
SELECT 
  id, 
  username, 
  email, 
  password_hash, 
  bio, 
  created_at, 
  updated_at
FROM users
WHERE id = $1;