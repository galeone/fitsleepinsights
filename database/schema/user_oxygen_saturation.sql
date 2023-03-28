create table if not exists oxygen_saturation(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    avg double precision not null,
    max double precision not null,
    min double precision not null
);