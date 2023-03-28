create table if not exists activity_calories_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists calories_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists calories_bmr_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists distance_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists elevation_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists floors_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists minutes_sedentary_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists minutes_lightly_active_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists minutes_fairly_active_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists minutes_very_active_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);

create table if not exists steps_series(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    value text not null
);