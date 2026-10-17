-- +goose Up

-- The service writes addresses already normalised; the check makes sure no
-- other writer can store a second spelling of one address.
create table accounts (
    id            uuid primary key,
    email         text not null,
    password_hash text not null,
    created_at    timestamptz not null,
    constraint accounts_email_key unique (email),
    constraint accounts_email_normalised check (email = lower(btrim(email)))
);

-- Only the SHA-256 of a token is kept, so reading this table signs no one in.
create table sessions (
    token_hash bytea primary key check (octet_length(token_hash) = 32),
    account_id uuid not null references accounts (id) on delete cascade,
    created_at timestamptz not null,
    expires_at timestamptz not null check (expires_at > created_at)
);
create index sessions_account_id_idx on sessions (account_id);
create index sessions_expires_at_idx on sessions (expires_at);

-- The checks repeat the domain's rules, as the last line of defence. An
-- account holds at most 20 alerts, so listing them needs no index beyond the
-- unique one, which starts with the account.
create table alerts (
    id            uuid primary key,
    account_id    uuid not null references accounts (id) on delete cascade,
    name          text not null check (name = btrim(name) and char_length(name) between 1 and 80),
    source        text not null check (source in ('wind', 'solar')),
    subsystem     text not null check (subsystem in ('N', 'NE', 'SE', 'S')),
    threshold_mwh integer not null check (threshold_mwh between 1 and 1000000),
    created_at    timestamptz not null
);
create unique index alerts_account_name_key on alerts (account_id, lower(name));

-- +goose Down
drop table alerts;
drop table sessions;
drop table accounts;
