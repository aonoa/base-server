create table users
(
    id          uuid                     not null
        primary key,
    create_time timestamp with time zone not null,
    update_time timestamp with time zone not null,
    username    varchar                  not null,
    password    varchar                  not null,
    nickname    varchar                  not null,
    status      smallint                 not null,
    avatar      varchar                  not null,
    "desc"      varchar                  not null,
    extension   varchar                  not null
);

comment on table users is 'Comment that appears in both the schema and the generated code';

comment on column users.username is '用户名';

comment on column users.password is '密码';

comment on column users.nickname is '昵称';

comment on column users.status is '0-锁定，1-正常';

comment on column users.avatar is '头像';

comment on column users."desc" is '备注';

comment on column users.extension is '扩展信息';

alter table users
    owner to postgres;

INSERT INTO public.users (id, create_time, update_time, username, password, nickname, status, avatar, "desc", extension) VALUES ('d1f7b7c1-c0b6-4707-aa17-5055b09b3ae8', '2023-05-17 14:29:18.185161 +00:00', '2023-06-03 15:20:19.378832 +00:00', 'vben', '123456', 'admin', 1, 'https://q1.qlogo.cn/g?b=qq&nk=190848757&s=640', 'test', 'AC_100100,AC_100110,AC_100120,AC_100010');
