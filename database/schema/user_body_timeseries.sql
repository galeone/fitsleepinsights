CREATE TABLE IF NOT EXISTS body_weight_series(
    id BIGSERIAL PRIMARY KEY,
    date_time timestamp without time zone not null,
    value DOUBLE PRECISION NOT NULL,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS bmi_series(
    id BIGSERIAL PRIMARY KEY,
    date_time timestamp without time zone not null,
    value DOUBLE PRECISION NOT NULL,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS body_fat_series(
    id BIGSERIAL PRIMARY KEY,
    date_time timestamp without time zone not null,
    value DOUBLE PRECISION NOT NULL,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id)
);