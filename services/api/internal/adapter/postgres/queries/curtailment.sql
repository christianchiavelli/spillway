-- name: FileVersions :many
select key, etag
from ons_files;

-- A plant keeps the name, state and subsystem of its latest record, so an
-- older month loaded later does not undo a rename.
-- name: UpsertPlant :batchexec
insert into plants (id, name, state, subsystem, source, last_seen)
values (@id, @name, @state, @subsystem, @source, @last_seen)
on conflict (id) do update
set name = excluded.name,
    state = excluded.state,
    subsystem = excluded.subsystem,
    source = excluded.source,
    last_seen = excluded.last_seen
where excluded.last_seen >= plants.last_seen;

-- name: DeletePlantDays :exec
delete from plant_days
where source = @source and day >= @first_day and day < @next_month;

-- name: DeleteSubsystemDays :exec
delete from subsystem_days
where source = @source and day >= @first_day and day < @next_month;

-- name: DeleteReasonDays :exec
delete from reason_days
where source = @source and day >= @first_day and day < @next_month;

-- name: DeleteHalfHours :exec
delete from half_hours
where source = @source and starts_at >= @from_instant and starts_at < @until_instant;

-- name: CopyPlantDays :copyfrom
insert into plant_days (day, plant_id, source, subsystem, curtailed_mwh, generated_mwh, limited_minutes)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: CopySubsystemDays :copyfrom
insert into subsystem_days (source, subsystem, day, curtailed_mwh, generated_mwh)
values ($1, $2, $3, $4, $5);

-- name: CopyReasonDays :copyfrom
insert into reason_days (day, source, reason, curtailed_mwh)
values ($1, $2, $3, $4);

-- name: CopyHalfHours :copyfrom
insert into half_hours (starts_at, source, curtailed_mwh)
values ($1, $2, $3);

-- name: RecordFile :exec
insert into ons_files (key, source, month, etag, records, last_instant, loaded_at)
values (@key, @source, @month, @etag, @records, sqlc.narg(last_instant), @loaded_at)
on conflict (key) do update
set etag = excluded.etag,
    records = excluded.records,
    last_instant = excluded.last_instant,
    loaded_at = excluded.loaded_at;

-- How far the data goes: from the first month of any source, to the latest
-- half hour of the source that is furthest behind, so a day is shown only
-- once both sources have it. No row until a file with records is loaded.
-- name: Bounds :one
select min(first_month)::date as first_day, min(latest)::timestamptz as latest
from (
    select min(month) filter (where records > 0) as first_month, max(last_instant) as latest
    from ons_files
    group by source
) s
having min(latest) is not null;

-- name: SourceTotals :many
select source, sum(curtailed_mwh)::float8 as curtailed_mwh, sum(generated_mwh)::float8 as generated_mwh
from subsystem_days
where day between @first_day and @last_day
group by source;

-- name: CurtailedBetween :one
select coalesce(sum(curtailed_mwh), 0)::float8 as curtailed_mwh
from subsystem_days
where day between @first_day and @last_day;

-- name: SubsystemTotals :many
select subsystem, source, sum(curtailed_mwh)::float8 as curtailed_mwh
from subsystem_days
where day between @first_day and @last_day
group by subsystem, source;

-- name: ReasonTotals :many
select reason, source, sum(curtailed_mwh)::float8 as curtailed_mwh
from reason_days
where day between @first_day and @last_day
group by reason, source;

-- name: DailySeries :many
select day, source, sum(curtailed_mwh)::float8 as curtailed_mwh
from subsystem_days
where day between @first_day and @last_day
group by day, source
order by day;

-- name: HalfHourSeries :many
select starts_at, source, curtailed_mwh
from half_hours
where starts_at >= @from_instant and starts_at < @until_instant and starts_at <= @latest
order by starts_at;

-- A plant is limited in a period when it lost any energy in it.
-- name: PlantCounts :one
select count(*) filter (where p.curtailed_mwh > 0) as limited, count(*) as reporting
from (
    select plant_id, sum(curtailed_mwh) as curtailed_mwh
    from plant_days
    where day between @first_day and @last_day
      and (sqlc.narg(source)::text is null or source = sqlc.narg(source))
      and (sqlc.narg(subsystem)::text is null or subsystem = sqlc.narg(subsystem))
    group by plant_id
) p;

-- name: RankPlants :many
select p.id, p.name, p.state, p.subsystem, p.source, t.curtailed_mwh, t.generated_mwh, t.limited_minutes
from (
    select plant_id,
           sum(curtailed_mwh)::float8 as curtailed_mwh,
           sum(generated_mwh)::float8 as generated_mwh,
           sum(limited_minutes)::bigint as limited_minutes
    from plant_days
    where day between @first_day and @last_day
      and (sqlc.narg(source)::text is null or source = sqlc.narg(source))
      and (sqlc.narg(subsystem)::text is null or subsystem = sqlc.narg(subsystem))
    group by plant_id
    having sum(curtailed_mwh) > 0
) t
join plants p on p.id = t.plant_id
order by t.curtailed_mwh desc, p.id
limit @row_limit offset @row_offset;

-- name: MonthlyTotals :many
select date_trunc('month', day)::date as month, source,
       sum(curtailed_mwh)::float8 as curtailed_mwh, sum(generated_mwh)::float8 as generated_mwh
from subsystem_days
where day <= @last_day
group by 1, 2
order by 1, 2;

-- name: SubsystemDaysBetween :many
select day, subsystem, curtailed_mwh, generated_mwh
from subsystem_days
where source = @source and day between @first_day and @last_day
order by day, subsystem;
