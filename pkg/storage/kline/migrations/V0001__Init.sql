create table if not exists kline
(
    open_time                    bigint primary key,
    open                         varchar,
    high                         varchar,
    low                          varchar,
    close                        varchar,
    volume                       varchar,
    close_time                   bigint,
    trade_num                    bigint
);