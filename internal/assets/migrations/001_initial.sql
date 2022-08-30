-- +migrate Up

create table native_deposits
(
    id             bigserial primary key not null,
    hash           text unique           not null,
    sender         text                  not null,
    receiver       text                  not null,
    target_network text                  not null,
    amount         bigint                not null,
);

create table ft_deposits
(
    id             bigserial primary key not null,
    hash           text unique           not null,
    sender         text                  not null,
    receiver       text                  not null,
    target_network text                  not null,
    amount         bigint                not null,
    mint           text                  not null,
);

create table nft_deposits
(
    id             bigserial primary key not null,
    hash           text unique           not null,
    sender         text                  not null,
    receiver       text                  not null,
    target_network text                  not null,
    mint           text                  not null,
    collection     text,
);

-- +migrate Down
drop table native_deposits;
drop table ft_deposits;
drop table nft_deposits;