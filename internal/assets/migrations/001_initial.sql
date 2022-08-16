-- +migrate Up

create table transactions
(
    id             bigserial primary key not null,
    hash           text                  not null,
    token_address  text                  not null,
    token_id       text                  not null,
    target_network text                  not null,
    receiver       text                  not null,
);

create index transactions_hash_index on transactions (hash);

-- +migrate Down
drop table transactions;