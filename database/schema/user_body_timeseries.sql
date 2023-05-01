CREATE TABLE IF NOT EXISTS body_weight_series(
    id bigserial primary key not null,
    date date not null,
    value double precision not null default 0,
    user_id bigint not null references oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS bmi_series(
    id bigserial primary key not null,
    date date not null,
    value double precision not null default 0,
    user_id bigint not null references oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS body_fat_series(
    id bigserial primary key not null,
    date date not null,
    value double precision not null default 0,
    user_id bigint not null references oauth2_authorized(id)
);