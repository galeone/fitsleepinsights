create table if not exists breathing_rate(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    breathing_rate double precision not null
);

create table if not exists cardio_fitness_score(
    id bigserial primary key,
    user_id bigint not null references oauth2_authorized(id),
    date_time timestamp without time zone not null,
    vo2_max double precision not null
);