CREATE TABLE IF NOT EXISTS heart_rate_variability_time_series(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    daily_rmssd double precision not null default 0,
    deep_rmssd double precision not null default 0
);