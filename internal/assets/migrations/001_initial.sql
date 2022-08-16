-- +migrate Up

create type status as ENUM ('pending', 'complete', 'failed');

create table vestings
(
    id      bigserial primary key not null,
    account varchar(44)           not null,
    seed    bytea                 not null,
    status  status                not null,
    date    timestamp without time zone
);

create index vestings_index on vestings (status, date, account, seed);

-- +migrate Down
drop table vestings;
drop index vestings_index;