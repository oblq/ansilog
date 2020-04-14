create table logs
(
    id serial not null,
    level smallint not null,
    message text not null,
    message_data jsonb not null,
    ts timestamp with time zone not null
);

create index logs_created_at_index
    on logs (ts);

create index logs_level_index
    on logs (level);
