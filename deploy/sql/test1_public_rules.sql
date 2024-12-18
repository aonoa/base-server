create table rules
(
    id          bigint generated by default as identity
        primary key,
    create_time timestamp with time zone not null,
    update_time timestamp with time zone not null,
    ptype       varchar                  not null,
    v0          varchar                  not null,
    v1          varchar                  not null,
    v2          varchar                  not null,
    v3          varchar                  not null,
    v4          varchar                  not null,
    v5          varchar                  not null
);

alter table rules
    owner to postgres;

