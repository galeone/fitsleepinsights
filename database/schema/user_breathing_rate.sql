CREATE TABLE IF NOT EXISTS breathing_rate_series(
    id BIGSERIAL primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone NOT NULL,
    breathing_rate FLOAT NOT NULL
);