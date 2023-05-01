-- /activities/heart/date/%s/%s.json

-- table heart_rate_zones created in user_activity.sql

CREATE TABLE IF NOT EXISTS heart_rate_activities(
    id bigserial primary key not null,
    date date not null,
    resting_heart_rate bigint not null,
    user_id bigint not null references oauth2_authorized(id)
);