CREATE TABLE IF NOT EXISTS weight_goals(
    id BIGSERIAL PRIMARY KEY not null,
    user_id bigint not null references oauth2_authorized(id),
    goal_type TEXT not null,
    start_date DATE not null,
    start_weight BIGINT not null,
    weight BIGINT not null,
    weight_threshold DOUBLE PRECISION not null
);

CREATE TABLE IF NOT EXISTS fat_goals(
    id bigserial primary key not null,
    fat BIGINT not null,
    user_id BIGINT not null REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS fat_logs(
    id bigserial primary key not null,
    fat BIGINT not null,
    log_id BIGINT not null,
    source TEXT not null,
    date_time timestamp without time zone not null,
    user_id BIGINT not null REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS weight_logs(
    id bigserial primary key not null,
    bmi DOUBLE PRECISION not null,
    fat BIGINT not null,
    log_id BIGINT not null,
    source TEXT not null,
    date_time timestamp without time zone not null,
    weight DOUBLE PRECISION not null,
    user_id BIGINT not null REFERENCES oauth2_authorized(id)
);