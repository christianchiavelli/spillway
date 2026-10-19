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

-- An alert fires on a day its source, in its subsystem, loses more than its
-- threshold; as much does not set it off. The rule lives in these three
-- queries, and an integration test holds them to it (ADR 0021).

-- Alerts that never fired have no row.
-- name: AlertActivity :many
select a.id, max(g.day)::date as last_fired_on, count(*) filter (where g.day >= @first_day) as days_fired
from alerts a
join subsystem_days g
  on g.source = a.source and g.subsystem = a.subsystem
 and g.day <= @last_day and g.curtailed_mwh > a.threshold_mwh
where a.account_id = @account_id
group by a.id;

-- name: CountFirings :one
select count(*)
from alerts a
join subsystem_days g on g.source = a.source and g.subsystem = a.subsystem
where a.account_id = @account_id
  and g.day between @first_day and @last_day
  and g.curtailed_mwh > a.threshold_mwh;

-- On a day with more than one, the furthest over its threshold first.
-- name: ListFirings :many
select a.id as alert_id, a.name as alert_name, g.day, g.curtailed_mwh, a.threshold_mwh
from alerts a
join subsystem_days g on g.source = a.source and g.subsystem = a.subsystem
where a.account_id = @account_id
  and g.day between @first_day and @last_day
  and g.curtailed_mwh > a.threshold_mwh
order by g.day desc, g.curtailed_mwh / a.threshold_mwh desc, a.id
limit @row_limit offset @row_offset;
