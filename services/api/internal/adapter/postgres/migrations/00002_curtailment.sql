-- +goose Up

-- The sums of the ONS's files, the ways the pages read them. Each file, one
-- source's month, is replaced whole when the ONS rewrites it, so every table
-- below carries the source, and a day, to delete it by. Energy is in MWh.

-- A plant, or a group of plants on one connection, as last seen.
create table plants (
    id        text primary key,
    name      text not null,
    state     text not null,
    subsystem text not null check (subsystem in ('N', 'NE', 'SE', 'S')),
    source    text not null check (source in ('wind', 'solar')),
    last_seen date not null
);

-- The plants ranked over a period: every plant's days in it, summed.
create table plant_days (
    day             date not null,
    plant_id        text not null references plants (id),
    source          text not null check (source in ('wind', 'solar')),
    subsystem       text not null check (subsystem in ('N', 'NE', 'SE', 'S')),
    curtailed_mwh   double precision not null check (curtailed_mwh >= 0),
    generated_mwh   double precision not null,
    limited_minutes integer not null check (limited_minutes >= 0),
    primary key (day, plant_id)
);

-- Totals, series, history and alerts: a source's day in a subsystem. Alerts
-- read it by source and subsystem, the rest by day.
create table subsystem_days (
    source        text not null check (source in ('wind', 'solar')),
    subsystem     text not null check (subsystem in ('N', 'NE', 'SE', 'S')),
    day           date not null,
    curtailed_mwh double precision not null check (curtailed_mwh >= 0),
    generated_mwh double precision not null,
    primary key (source, subsystem, day)
);
create index subsystem_days_day_idx on subsystem_days (day);

create table reason_days (
    day           date not null,
    source        text not null check (source in ('wind', 'solar')),
    reason        text not null check (reason in ('ENE', 'CNF', 'REL', 'PAR')),
    curtailed_mwh double precision not null check (curtailed_mwh >= 0),
    primary key (day, source, reason)
);

-- A day's chart, the whole grid.
create table half_hours (
    starts_at     timestamptz not null,
    source        text not null check (source in ('wind', 'solar')),
    curtailed_mwh double precision not null check (curtailed_mwh >= 0),
    primary key (starts_at, source)
);

-- The files loaded, and the ETag each had: a file is read again only when its
-- ETag changes. last_instant is the file's last half hour, null when empty.
create table ons_files (
    key          text primary key,
    source       text not null check (source in ('wind', 'solar')),
    month        date not null check (extract(day from month) = 1),
    etag         text not null,
    records      integer not null,
    last_instant timestamptz,
    loaded_at    timestamptz not null,
    unique (source, month)
);

-- +goose Down
drop table ons_files;
drop table half_hours;
drop table reason_days;
drop table subsystem_days;
drop table plant_days;
drop table plants;
