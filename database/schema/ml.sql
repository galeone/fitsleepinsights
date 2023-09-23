create table if not exists predictors(
    id bigserial primary key not null,
    user_id bigint not null references oauth2_authorized(id),
    target text not null,
    endpoint text not null,
    created_at timestamp not null default now()
);