CREATE TABLE IF NOT EXISTS weight_goals(
    id BIGSERIAL PRIMARY KEY not null,
    user_id bigint not null references oauth2_authorized(id),
    goal_type TEXT NOT NULL,
    start_date DATE NOT NULL,
    start_weight BIGINT NOT NULL,
    weight BIGINT NOT NULL,
    weight_threshold DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS fat_goals(
    id BIGSERIAL PRIMARY KEY,
    fat BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS fat_logs(
    id BIGSERIAL PRIMARY KEY,
    fat BIGINT NOT NULL,
    log_id BIGINT NOT NULL,
    source TEXT NOT NULL,
    date_time timestamp without time zone not null,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id)
);

CREATE TABLE IF NOT EXISTS weight_logs(
    id BIGSERIAL PRIMARY KEY,
    bmi DOUBLE PRECISION NOT NULL,
    fat BIGINT NOT NULL,
    log_id BIGINT NOT NULL,
    source TEXT NOT NULL,
    date_time timestamp without time zone not null,
    weight DOUBLE PRECISION NOT NULL,
    user_id BIGINT NOT NULL REFERENCES oauth2_authorized(id)
);