CREATE TABLE IF NOT EXISTS body_weight_series(
    id bigserial primary key not null,
    date_time timestamp without time zone not null,
    value DOUBLE PRECISION not null,
    user_id BIGINT not null REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS bmi_series(
    id bigserial primary key not null,
    date_time timestamp without time zone not null,
    value DOUBLE PRECISION not null,
    user_id BIGINT not null REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS body_fat_series(
    id bigserial primary key not null,
    date_time timestamp without time zone not null,
    value DOUBLE PRECISION not null,
    user_id BIGINT not null REFERENCES oauth2_authorized(id)
);