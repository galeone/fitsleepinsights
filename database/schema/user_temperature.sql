create table if not exists core_temperatures(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    value double precision not null
);

create table if not exists skin_temperatures(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    value double precision not null,
    log_type text not null
);