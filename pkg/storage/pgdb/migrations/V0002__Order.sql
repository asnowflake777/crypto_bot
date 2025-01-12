create table users
(
    uid   serial primary key,
    login varchar unique
);

create table orders
(
    id       serial primary key,
    symbol   varchar,
    price    float8,
    quantity float8,
    type     varchar,
    side     varchar,
    user_uid integer references users (uid) on delete cascade
);

create table balance
(
    asset    varchar,
    free     float8,
    locked   float8,
    user_uid integer references users (uid) on delete cascade,

    primary key (asset, user_uid)
);
