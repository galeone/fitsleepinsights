-- /activities.json
create table if not exists categories(
    id bigserial primary key not null,
    name text not null,
    unique(name)
);

create table if not exists subcategories(
    id bigint primary key not null,
    name text not null,
    category bigint references categories(id),
    unique(name)
);

create table if not exists activities_descriptions(
    id bigint primary key not null,
    access_level text not null,
    has_speed bool not null default false,
    mets double precision not null default 0,
    name text not null,
    subcategory bigint references subcategories(id),
    category bigint references categories(id)
);

create table if not exists activity_levels(
    id bigint primary key not null,
    max_speed_mph double precision not null default 0,
    min_speed_mph double precision not null default 0,
    mets double precision not null default 0,
    name text not null,
    activity_description bigint references activities_descriptions(id)
);