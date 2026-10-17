-- Taken before counting, so two inserts for one account run one after the
-- other. NO KEY UPDATE, because nothing here changes the key: inserts that
-- only reference the account, like a new session, are not held up.
-- name: LockAccount :one
select id
from accounts
where id = $1
for no key update;

-- name: CountAlerts :one
select count(*)
from alerts
where account_id = $1;

-- name: InsertAlert :exec
insert into alerts (id, account_id, name, source, subsystem, threshold_mwh, created_at)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAlerts :many
select id, account_id, name, source, subsystem, threshold_mwh, created_at
from alerts
where account_id = $1
order by created_at desc, id desc;

-- name: DeleteAlert :execrows
delete from alerts
where id = $1 and account_id = $2;
