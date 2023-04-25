create table if not exists sleep_logs(
    log_id bigint not null primary key, -- no bigserial, server side unique
    user_id bigint not null references oauth2_authorized(id),
    date_of_sleep date not null,
    duration bigint not null default 0,
    efficiency bigint not null default 0,
    end_time timestamp without time zone not null,
    info_code bigint not null default 0,
    is_main_sleep boolean not null default false,
    log_type text not null,
    minutes_after_wakeup bigint not null default 0,
    minutes_asleep bigint not null default 0,
    minutes_awake bigint not null default 0,
    minutes_to_fall_asleep bigint not null default 0,
    start_time timestamp without time zone not null,
    time_in_bed bigint not null default 0,
    "type" text not null
);

create table if not exists sleep_stage_details(
    id bigserial primary key not null,
    count bigint not null default 0,
    minutes bigint not null default 0,
    thirty_day_avg_minutes bigint not null default 0,
    sleep_log_id bigint not null references sleep_logs(log_id)
);

create table if not exists sleep_data(
    id bigserial primary key not null,
    date_time timestamp without time zone not null,
    level text not null,
    sleep_log_id bigint not null references sleep_logs(log_id),
    seconds bigint not null default 0
);

create table if not exists sleep_stages_summary(
    id bigserial primary key not null,
    deep bigint not null default 0,
    light bigint not null default 0,
    rem bigint not null default 0,
    wake bigint not null default 0
);

create table if not exists sleep_summary(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    stages_id bigint references sleep_stages_summary(id),
    total_minutes_asleep bigint not null default 0,
    total_sleep_records bigint not null default 0,
    total_time_in_bed bigint not null default 0
);