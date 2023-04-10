-- /activities/heart/date/%s/%s.json

-- table heart_rate_zones created in user_activity.sql

CREATE TABLE IF NOT EXISTS heart_rate_time_point_values(
    id bigserial PRIMARY KEY not null,
    resting_heart_rate bigint not null default 0,
    heart_rate_zone_id bigint references heart_rate_zones(id)
);

CREATE TABLE IF NOT EXISTS heart_rate_activities(
    id bigserial primary key not null,
    date_time timestamp without time zone not null,
    user_id bigint not null references oauth2_authorized(id),
    heart_rate_time_point_value bigint references heart_rate_time_point_values(id)
);