CREATE TABLE IF NOT EXISTS heart_rate_variability_time_series(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    daily_rmssd double precision not null,
    deep_rmssd double precision not null
);