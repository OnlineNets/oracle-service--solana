-- +migrate Up

create table native_deposits
(
    id             bigserial primary key not null,
    hash           text                  not null,
    instruction_id integer               not null,
    sender         text                  not null,
    receiver       text                  not null,
    target_network text                  not null,
    amount         bigint                not null,
    bundle_data    text,
    bundle_seed    text
);

create unique index native_index on native_deposits (hash, instruction_id);

create table ft_deposits
(
    id             bigserial primary key not null,
    hash           text                  not null,
    instruction_id integer               not null,
    sender         text                  not null,
    receiver       text                  not null,
    target_network text                  not null,
    amount         bigint                not null,
    mint           text                  not null,
    bundle_data    text,
    bundle_seed    text
);

create unique index ft_index on ft_deposits (hash, instruction_id);

create table nft_deposits
(
    id             bigserial primary key not null,
    hash           text                  not null,
    instruction_id integer               not null,
    sender         text                  not null,
    receiver       text                  not null,
    target_network text                  not null,
    mint           text                  not null,
    collection     text,
    bundle_data    text,
    bundle_seed    text
);

create unique index nft_index on nft_deposits (hash, instruction_id);

-- +migrate Down
drop table native_deposits;
drop table ft_deposits;
drop table nft_deposits;