-- /activities/goals/%s.json
create table if not exists goals(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    active_minutes bigint not null,
    calories_out bigint not null,
    distance double precision not null,
    steps bigint not null,
    -- manually added (period: weekly, daily)
    start_date date not null,
    end_date date not null
);

-- /activities/list.json?afterDate=2022-10-29&sort=asc&offset=0&limit=2
/*
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
*/

create table if not exists manual_values_specified(
    id bigserial primary key not null,
    calories bool not null default false,
    distance bool not null default false,
    steps bool not null default false
);

create table if not exists log_sources(
    -- ID is not a big serial because this is the ID of the device
    -- that's unique and sent by the API.
    id bigint primary key not null,
    -- tracker_features fitbit_features [] not null,
    tracker_features text[] not null,
    name text not null,
    "type" text not null,
    url text not null
);

create table if not exists active_zone_minutes(
    id bigserial primary key not null,
    total_minutes bigint not null
);

create table if not exists minutes_in_heart_rate_zone(
    id bigserial primary key not null,
    active_zone_minutes_id bigint not null references active_zone_minutes(id),
    minute_multiplier bigint not null,
    minutes bigint not null,
    "order" bigint not null,
    "type" text not null,
    zone_name text not null
);

create table if not exists activity_logs(
    -- where the default is 0 (or equivalent) is because
    -- there could be activities without these values
    -- e.g sedentary = 0 distance, thus useless distance unit, 0 pace, 0 speed, ...
    log_id bigint primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    active_duration bigint not null,
    active_zone_minutes_id bigint not null references active_zone_minutes(id),
    activity_name text not null,
    activity_type_id bigint not null,
    average_heart_rate bigint not null,
    calories bigint not null,
    distance double precision not null default 0,
    distance_unit text not null default 'nd',
    duration bigint not null,
    elevation_gain bigint not null default 0,
    has_active_zone_minutes bool not null default false,
    heart_rate_link text not null,
    last_modified text not null,
    log_type text not null,
    manual_values_specified_id bigint not null references manual_values_specified(id),
    original_duration bigint not null,
    original_start_time timestamp not null,
    pace double precision not null default 0,
    source_id bigint references log_sources(id), --nullable
    speed double precision not null default 0,
    start_time timestamp not null,
    steps bigint not null default 0,
    tcx_link text not null
);

/*
 do $$
 begin create type heart_rate_zone_type as enum ('CUSTOM', 'DEFAULT');
 EXCEPTION
 WHEN duplicate_object THEN null;
 end $$;
 */
create table if not exists heart_rate_zones(
    id bigserial primary key not null,
    -- nullable activity_log_id because heart_rate_zones is also used in
    -- user_hr_timeseries that's not connected to an activity_log
    activity_log_id bigint null references activity_logs(log_id),
    calories_out double precision not null,
    max bigint not null,
    min bigint not null,
    minutes bigint not null,
    name text not null,
    "type" text null default 'DEFAULT'
);

create table if not exists logged_activity_levels(
    id bigserial primary key not null,
    activity_log_id bigint not null references activity_logs(log_id),
    minutes bigint not null,
    name text not null
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
    id bigserial primary key not null,
    activities_summary_id bigint not null references activities_summaries(id),
    distance_id bigint not null references distances(id),
    unique (activities_summary_id, distance_id)
);

create table if not exists activities_summary_heart_rate_zones(
    id bigserial primary key not null,
    activities_summary_id bigint not null references activities_summaries(id),
    heart_rate_zone_id bigint not null references heart_rate_zones(id),
    unique (activities_summary_id, heart_rate_zone_id)
);

create table if not exists daily_activity_summaries(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    goals_id bigint not null references goals(id),
    summary_id bigint not null references activities_summaries(id)
);

create table if not exists daily_activity_summary_activities(
    id bigserial primary key not null,
    daily_activity_summary_id bigint not null references daily_activity_summaries(id),
    activities_summary_id bigint not null references activities_summaries(id),
    unique (daily_activity_summary_id, activities_summary_id)
);

-- /activities.json
create table if not exists life_time_time_steps(
    id bigserial primary key not null,
    date date not null,
    value double precision not null
);

create table if not exists life_time_activities(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    distance_id bigint not null references life_time_time_steps(id),
    steps_id bigint not null references life_time_time_steps(id),
    floors_id bigint not null references life_time_time_steps(id)
);

create table if not exists life_time_stats(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    active_score bigint not null,
    calories_out bigint not null,
    distance double precision not null,
    steps bigint not null,
    floors bigint not null
);

create table if not exists best_stats_sources(
    id bigserial primary key not null,
    total_id bigint not null references life_time_activities(id),
    tracker_id bigint not null references life_time_activities(id)
);

create table if not exists lifetime_stats_sources(
    id bigserial primary key not null,
    total_id bigint not null references life_time_stats(id),
    tracker_id bigint not null references life_time_stats(id)
);

create table if not exists user_life_time_stats(
    id bigserial primary key not null,
    best_id bigint not null references best_stats_sources(id),
    lifetime_id bigint not null references lifetime_stats_sources(id)
);

-- /activities/favorite.json
create table if not exists favorite_activities(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    activity_id bigint not null,
    description text not null,
    mets bigint not null,
    name text not null
);

create table if not exists minimal_activities(
    id bigserial primary key not null,
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
    id bigserial primary key not null,
    minimal_activity_id bigint not null references minimal_activities(id)
);

-- /activities/recent.json
create table if not exists recent_activities(
    id bigserial primary key not null,
    minimal_activity_id bigint not null references minimal_activities(id)
);
