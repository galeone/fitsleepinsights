create table if not exists oxygen_saturation(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    avg double precision not null default 0,
    max double precision not null default 0,
    min double precision not null default 0
);