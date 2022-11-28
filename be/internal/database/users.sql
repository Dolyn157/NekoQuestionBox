drop table if exists public.users;
create table if not exists public.users(
    username text primary key,
    pwd text
);