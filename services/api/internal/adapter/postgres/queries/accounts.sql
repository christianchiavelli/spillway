-- name: CreateAccount :exec
insert into accounts (id, email, password_hash, created_at)
values ($1, $2, $3, $4);

-- name: AccountByEmail :one
select id, email, password_hash, created_at
from accounts
where email = $1;
