create table if not exists calories_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null default 0,
    mets bigint not null default 0,
    value double precision not null default 0
);

create table if not exists distance_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null default 0,
    mets bigint not null default 0,
    value double precision not null default 0
);

create table if not exists elevation_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null default 0,
    mets bigint not null default 0,
    value double precision not null default 0
);

create table if not exists floors_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null default 0,
    mets bigint not null default 0,
    value double precision not null default 0
);

create table if not exists steps_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null default 0,
    mets bigint not null default 0,
    value double precision not null default 0
);

create table if not exists oxygen_saturation_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    value double precision not null default 0
);

create table if not exists heart_rate_variability_intraday_hrv(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    coverage double precision not null default 0,
    hf double precision not null default 0,
    lf double precision not null default 0,
    rmssd double precision not null default 0
);

create table if not exists breathing_rate_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    deep_sleep_summary double precision not null default 0,
    full_sleep_summary double precision not null default 0,
    light_sleep_summary double precision not null default 0,
    rem_sleep_summary double precision not null default 0
);