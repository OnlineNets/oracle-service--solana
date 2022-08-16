-- +migrate Up

create table transactions
(
    id             bigserial primary key not null,
    hash           text                  not null,
    collection     text unique           not null,
    token_mint     text                  not null,
    token_id       text,
    target_network text                  not null,
    receiver       text                  not null
);

-- +migrate Down
drop table transactions;