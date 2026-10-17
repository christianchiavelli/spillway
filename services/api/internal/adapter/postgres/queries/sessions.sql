-- name: CreateSession :exec
insert into sessions (token_hash, account_id, created_at, expires_at)
values ($1, $2, $3, $4);

-- name: FindSession :one
select s.token_hash, s.account_id, s.created_at, s.expires_at,
       a.email, a.created_at as account_created_at
from sessions s
join accounts a on a.id = s.account_id
where s.token_hash = $1;

-- name: ExtendSession :exec
update sessions
set expires_at = $2
where token_hash = $1;

-- name: DeleteSession :exec
delete from sessions
where token_hash = $1;

-- name: DeleteExpiredSessions :execrows
delete from sessions
where expires_at <= $1;
