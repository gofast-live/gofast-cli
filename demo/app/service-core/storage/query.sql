-- name: Ping :one
select 1;

-- name: SelectToken :one
select * from tokens where id = $1;

-- name: InsertToken :one
insert into tokens (id, expires, target, callback) values ($1, $2, $3, $4) returning *;

-- name: UpdateToken :exec
update tokens set expires = $1 where id = $2 returning *;

-- name: DeleteTokens :exec
delete from tokens where expires < current_timestamp;

-- name: SelectUsers :many
select * from users;

-- name: SelectUser :one
select * from users where id = $1;

-- name: SelectUserByCustomerID :one
select * from users where customer_id = $1;

-- name: SelectUserByEmailAndSub :one
select * from users where email = $1 and sub = $2;

-- name: InsertUser :one
insert into users (id, email, access, sub, avatar) values ($1, $2, $3, $4, $5) returning *;

-- name: UpdateUserPhone :exec
update users set phone = $2 where id = $1;

-- name: UpdateUserActivity :exec
update users set updated = current_timestamp where id = $1;

-- name: UpdateUserCustomerID :exec
update users set customer_id = $1 where id = $2;

-- name: UpdateUserSubscription :exec
update users set access = $1, subscription_id = $2, subscription_end = $3 where customer_id = $4;

-- name: UpdateUserAccess :one
update users set access = $1 where id = $2 returning *;

-- name: CountNotes :one
select count(*) from notes where user_id = $1;

-- name: SelectNotes :many
select * from notes where user_id = $1 order by created desc limit $2 offset $3;

-- name: SelectNote :one
select * from notes where id = $1;

-- name: InsertNote :one
insert into notes (id, user_id, title, category, content) values ($1, $2, $3, $4, $5) returning *;

-- name: UpdateNote :one
update notes set title = $1, category = $2, content = $3 where id = $4 returning *;

-- name: DeleteNote :exec
delete from notes where id = $1;
