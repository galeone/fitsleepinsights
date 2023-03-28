create table if not exists calories_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null,
    mets bigint not null,
    value double precision not null
);

create table if not exists distance_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null,
    mets bigint not null,
    value double precision not null
);

create table if not exists elevation_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null,
    mets bigint not null,
    value double precision not null
);

create table if not exists floors_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null,
    mets bigint not null,
    value double precision not null
);

create table if not exists steps_series_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    level bigint not null,
    mets bigint not null,
    value double precision not null
);

create table if not exists oxygen_saturation_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    value double precision not null
);

create table if not exists heart_rate_variability_intraday_hrv(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    coverage double precision not null,
    hf double precision not null,
    lf double precision not null,
    rmssd double precision not null
);

create table if not exists breathing_rate_intraday(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    deep_sleep_summary double precision not null,
    full_sleep_summary double precision not null,
    light_sleep_summary double precision not null,
    rem_sleep_summary double precision not null
);