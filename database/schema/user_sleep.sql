create table if not exists sleep_stage_details(
    id bigserial primary key not null,
    count bigint not null,
    minutes bigint not null,
    thirty_day_avg_minutes bigint not null
);

create table if not exists sleep_data(
    id bigserial primary key not null,
    date_time timestamp without time zone not null,
    level text not null,
    seconds bigint not null
);

create table if not exists sleep_levels(
    id bigserial primary key not null,
    sleep_level_id bigint references sleep_data(id),
    short_data_id bigint references sleep_data(id),
    summary_id bigint references sleep_stage_details(id)
);

create table if not exists sleep_logs(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_of_sleep date not null,
    duration bigint not null,
    efficiency bigint not null,
    end_time text not null,
    info_code bigint not null,
    is_main_sleep boolean not null,
    levels_id bigint references sleep_levels(id),
    log_id bigint not null,
    log_type text not null,
    minutes_after_wakeup bigint not null,
    minutes_asleep bigint not null,
    minutes_awake bigint not null,
    minutes_to_fall_asleep bigint not null,
    start_time text not null,
    time_in_bed bigint not null,
    "type" text not null
);

create table if not exists sleep_stages_summary(
    id bigserial primary key not null,
    deep bigint not null,
    light bigint not null,
    rem bigint not null,
    wake bigint not null
);

create table if not exists sleep_summary(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    stages_id bigint references sleep_stages_summary(id),
    total_minutes_asleep bigint not null,
    total_sleep_records bigint not null,
    total_time_in_bed bigint not null
);