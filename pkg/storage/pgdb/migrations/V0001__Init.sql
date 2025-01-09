create table if not exists kline
(
    open_time                    bigint primary key,
    open                         float8,
    high                         float8,
    low                          float8,
    close                        float8,
    volume                       float8,
    close_time                   bigint,
    trade_num                    bigint
);