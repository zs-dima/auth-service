-- name: GetActiveUser :one
SELECT * FROM "user"
 WHERE email = $1 
   AND deleted_at IS NULL OR deleted_at > NOW()
 LIMIT 1;

-- name: LoadUsers :many
SELECT id,
       role,
       name,
       email,
       blurhash,
       deleted_at IS NOT NULL AND deleted_at < NOW() AS deleted
  FROM "user"
 ORDER BY name;

-- name: CreateUser :one
INSERT INTO "user" (
  id,
  role,
  name,
  email,
  password,
  deleted_at
) VALUES (
  $1, $2, $3, $4, $5,
  CASE WHEN sqlc.arg('Deleted')::bool THEN NOW() ELSE NULL END
)
RETURNING id;

-- name: UpdateUser :exec
UPDATE "user"
   SET role = $2,
       name = $3,
      email = $4,
 deleted_at = CASE WHEN sqlc.arg('Deleted')::bool THEN NOW() ELSE NULL END
 WHERE id = $1;

-- name: UpdateUserBlurhash :exec
UPDATE "user"
   SET blurhash = $2
 WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE "user"
   SET password = $2
 WHERE email = $1;

-- name: DeleteUser :exec
UPDATE "user"
   SET deleted_at = NOW()
 WHERE id = $1;


-- name: SaveUserPhoto :exec
INSERT INTO user_photo (
  user_id,
  avatar,
  photo
)
VALUES ($1, $2, $3)
ON CONFLICT(user_id) DO UPDATE SET 
    avatar = excluded.avatar,
    photo = excluded.photo;

-- name: LoadUsersAvatars :many
SELECT user_id, avatar
  FROM user_photo;

-- name: LoadUserAvatar :many
SELECT user_id, avatar 
  FROM user_photo 
 WHERE user_id = ANY($1::UUID[]);


-- name: SaveUserSession :exec
INSERT INTO user_session (
  user_id,
  refresh_token,
  expires_at
)
VALUES ($1, $2, $3)
ON CONFLICT(user_id) DO UPDATE SET 
    refresh_token = excluded.refresh_token,
    expires_at = excluded.expires_at;

-- name: EndUserSession :exec
UPDATE user_session
   SET refresh_token = '',
       expires_at = NOW()
 WHERE user_id = $1;

-- name: LoadRefreshToken :one
SELECT refresh_token
  FROM user_session
 WHERE user_id = $1 
   AND deleted_at is null;

