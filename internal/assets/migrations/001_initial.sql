-- +migrate Up

create table transactions
(
    hash           text primary key not null,
    token_address  text             not null,
    token_id       text             not null,
    target_network text             not null,
    receiver       text             not null,
);

-- +migrate Down
drop table transactions;