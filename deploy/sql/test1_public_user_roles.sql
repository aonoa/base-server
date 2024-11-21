create table user_roles
(
    user_id uuid   not null
        constraint user_roles_user_id
            references users
            on delete cascade,
    role_id bigint not null
        constraint user_roles_role_id
            references roles
            on delete cascade,
    primary key (user_id, role_id)
);

alter table user_roles
    owner to postgres;

