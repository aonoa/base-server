create table dept_users
(
    dept_id bigint not null
        constraint dept_users_dept_id
            references depts
            on delete cascade,
    user_id uuid   not null
        constraint dept_users_user_id
            references users
            on delete cascade,
    primary key (dept_id, user_id)
);

alter table dept_users
    owner to postgres;

