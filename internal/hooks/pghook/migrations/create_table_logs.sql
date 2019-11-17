create table logs (
    id serial not null,
    level smallint not null,
    message text not null,
    message_data json not null,
    created_at timestamp with time zone not null
);

create index logs_level_index on logs (level);

create index logs_created_at_index on logs (created_at);
