-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create type public.user_role as enum ('administrator', 'user');
alter type public.user_role owner to admin;

create table if not exists public."user"
(
    id uuid default uuid_generate_v4() primary key,
    role            user_role            not null,
    name            varchar(256)         not null,
    email           varchar(256)         not null
        constraint email
            unique,
    password        varchar(512)         not null,
    blurhash        varchar(37),
    deleted_at      timestamp
);
create index user_email_idx on public."user"(email);


create table if not exists public.user_session
(
    user_id         uuid not null primary key
        constraint user_session_user_id_fk
            references public."user",
    refresh_token   varchar(2048) not null,
    expires_at      timestamp     not null
);
create index user_session_expires_at_idx on public.user_session(expires_at);


create table if not exists public.user_photo
(
    user_id    uuid not null primary key
        constraint user_photo_user_id_fk
            references public."user",
    avatar     bytea,
    photo      bytea
);
