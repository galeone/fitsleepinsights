-- /activities.json
create table if not exists categories(
    id bigint not null primary key,
    name text not null
);

create table if not exists subcategories(
    id bigint not null primary key,
    name text not null,
    category bigint references categories(id)
);

create table if not exists activities_descriptions(
    id bigint not null primary key,
    access_level text not null,
    has_speed bool not null,
    mets bigint not null,
    name text not null,
    subcategory bigint references subcategories(id),
    category bigint references categories(id)
);

create table if not exists activity_levels(
    id bigint not null primary key,
    max_speed_mph double precision not null,
    min_speed_mph double precision not null,
    mets bigint not null,
    name text not null,
    activity_description bigint references activities_descriptions(id)
);