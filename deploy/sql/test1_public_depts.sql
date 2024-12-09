create table depts
(
    id          bigint generated by default as identity
        primary key,
    create_time timestamp with time zone not null,
    update_time timestamp with time zone not null,
    name        varchar                  not null,
    sort        bigint                   not null,
    status      boolean                  not null,
    "desc"      varchar                  not null,
    extension   varchar                  not null,
    dom         bigint                   not null,
    dept_roles  bigint
        constraint depts_roles_roles
            references roles
            on delete set null,
    pid         bigint
        constraint depts_depts_children
            references depts
            on delete set null
);

comment on column depts.name is '部门名称';

comment on column depts.sort is '排序';

comment on column depts.status is '0-锁定，1-正常';

comment on column depts."desc" is '备注';

comment on column depts.extension is '扩展信息';

comment on column depts.dom is '域';

comment on column depts.pid is '父节点id';

alter table depts
    owner to postgres;
