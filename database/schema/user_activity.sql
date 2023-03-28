-- /activities/goals/%s.json
create table if not exists goals(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    active_minutes bigint not null,
    calories_out bigint not null,
    distance double precision not null,
    steps bigint not null
);

-- /activities/list.json?afterDate=2022-10-29&sort=asc&offset=0&limit=2
do $$
begin create type fitbit_features as enum(
    'CALORIES',
    'DISTANCE',
    'ELEVATION',
    'GPS',
    'HEARTRATE',
    'PACE',
    'STEPS',
    'VO2_MAX'
);
EXCEPTION
WHEN duplicate_object THEN null;
end $$;

create table if not exists manual_values_specified(
    id bigserial primary key not null,
    calories bool not null,
    distance bool not null,
    steps bool not null
);

do $$
begin create type heart_rate_zone_type as enum ('CUSTOM', 'DEFAULT');
EXCEPTION
WHEN duplicate_object THEN null;
end $$;

create table if not exists heart_rate_zones(
    id bigserial primary key not null,
    calories_out double precision not null,
    max bigint not null,
    min bigint not null,
    minutes bigint not null,
    name text not null,
    type heart_rate_zone_type not null default 'DEFAULT'::heart_rate_zone_type
);

create table if not exists log_sources(
    id bigserial primary key not null,
    source_id text not null,
    name text not null,
    type text not null,
    url text not null
);

create table if not exists minutes_in_heart_rate_zone(
    id bigserial primary key not null,
    minute_multiplier bigint not null,
    minutes bigint not null,
    "order" bigint not null,
    "type" text not null,
    zone_name text not null
);

create table if not exists logged_activity_levels(
    id bigserial primary key not null,
    minutes bigint not null,
    name text not null
);

create table if not exists active_zone_minutes(
    id bigserial primary key not null,
    total_minutes bigint not null
);

create table if not exists minutes_in_heart_rate_zones_list(
    active_zone_minutes_id integer not null references active_zone_minutes(id),
    minutes_in_heart_rate_zone_id integer not null references minutes_in_heart_rate_zone(id),
    primary key(
        active_zone_minutes_id,
        minutes_in_heart_rate_zone_id
    )
);

create table if not exists activity_logs(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    active_duration bigint not null,
    active_zone_minutes_id integer not null references active_zone_minutes(id),
    source_tracker_features fitbit_features [] not null,
    activity_name text not null,
    activity_type_id bigint not null,
    average_heart_rate bigint not null,
    calories bigint not null,
    distance double precision not null,
    distance_unit text not null,
    duration bigint not null,
    elevation_gain bigint not null,
    has_active_zone_minutes bool not null,
    heart_rate_link text not null,
    last_modified text not null,
    log_id bigint not null,
    log_type text not null,
    manual_values_specified_id integer not null references manual_values_specified(id),
    original_duration bigint not null,
    original_start_time timestamp not null,
    pace double precision not null,
    source_id integer not null references log_sources(id),
    speed double precision not null,
    start_time timestamp not null,
    steps bigint not null,
    tcx_link text not null
);

create table if not exists activity_log_heart_rate_zones(
    activity_log_id integer not null references activity_logs(id),
    heart_rate_zone_id integer not null references heart_rate_zones(id),
    primary key (activity_log_id, heart_rate_zone_id)
);

create table if not exists activity_log_activity_levels(
    activity_log_id integer not null references activity_logs(id),
    logged_activity_level_id integer not null references logged_activity_levels(id),
    primary key (activity_log_id, logged_activity_level_id)
);

-- /activities/date/%s.json
create table if not exists distances(
    id bigserial primary key not null,
    activity text not null,
    distance double precision not null
);

create table if not exists activities_summaries(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    active_score bigint not null,
    activity_calories bigint not null,
    calories_bmr bigint not null,
    calories_out bigint not null,
    fairly_active_minutes bigint not null,
    lightly_active_minutes bigint not null,
    marginal_calories bigint not null,
    resting_heart_rate bigint not null,
    sedentary_minutes bigint not null,
    steps bigint not null,
    very_active_minutes bigint not null
);

create table if not exists activities_summary_distances(
    activities_summary_id integer not null references activities_summaries(id),
    distance_id integer not null references distances(id),
    primary key (activities_summary_id, distance_id)
);

create table if not exists activities_summary_heart_rate_zones(
    activities_summary_id integer not null references activities_summaries(id),
    heart_rate_zone_id integer not null references heart_rate_zones(id),
    primary key (activities_summary_id, heart_rate_zone_id)
);

create table if not exists daily_activity_summaries(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    goals_id integer not null references goals(id),
    summary_id integer not null references activities_summaries(id)
);

create table if not exists daily_activity_summary_activities(
    daily_activity_summary_id integer not null references daily_activity_summaries(id),
    activities_summary_id integer not null references activities_summaries(id),
    primary key (daily_activity_summary_id, activities_summary_id)
);

-- /activities.json
create table if not exists life_time_time_steps(
    id bigserial primary key,
    date date not null,
    value double precision not null
);

create table if not exists life_time_activities(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    distance_id integer not null references life_time_time_steps(id),
    steps_id integer not null references life_time_time_steps(id),
    floors_id integer not null references life_time_time_steps(id)
);

create table if not exists life_time_stats(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    active_score bigint not null,
    calories_out bigint not null,
    distance double precision not null,
    steps bigint not null,
    floors bigint not null
);

create table if not exists best_stats_sources(
    id bigserial primary key,
    total_id integer not null references life_time_activities(id),
    tracker_id integer not null references life_time_activities(id)
);

create table if not exists lifetime_stats_sources(
    id bigserial primary key,
    total_id integer not null references life_time_stats(id),
    tracker_id integer not null references life_time_stats(id)
);

create table if not exists user_life_time_stats(
    id bigserial primary key,
    best_id integer not null references best_stats_sources(id),
    lifetime_id integer not null references lifetime_stats_sources(id)
);

-- /activities/favorite.json
create table if not exists favorite_activities(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    activity_id bigint not null,
    description text not null,
    mets bigint not null,
    name text not null
);

create table if not exists minimal_activities(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    activity_id bigint not null,
    calories bigint not null,
    description text not null,
    distance double precision not null,
    duration bigint not null,
    name text not null
);

-- /activities/frequent.json
create table if not exists frequent_activities(
    id bigserial primary key,
    minimal_activity_id integer not null references minimal_activities(id)
);

-- /activities/recent.json
create table if not exists recent_activities(
    id bigserial primary key,
    minimal_activity_id integer not null references minimal_activities(id)
);