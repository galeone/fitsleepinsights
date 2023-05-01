create table if not exists breathing_rate(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    breathing_rate double precision not null default 0
);

create table if not exists cardio_fitness_score(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    date date not null,
    vo2max_lower_bound double precision not null default 0,
    vo2max_upper_bound double precision not null default 0
);