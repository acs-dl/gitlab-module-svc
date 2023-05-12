-- +migrate Up

create table if not exists responses (
    id uuid primary key,
    status text not null,
    error text,
    payload jsonb not null,
    created_at timestamp without time zone not null default current_timestamp
);

create table if not exists users (
    id bigint unique,
    username text not null unique,
    gitlab_id bigint primary key,
    avatar_url text not null,
    name text not null,
    updated_at timestamp with time zone not null default current_timestamp,
    created_at timestamp with time zone default current_timestamp
);

create index if not exists users_id_idx on users(id);
create index if not exists users_username_idx on users(username);
create index if not exists users_gitlabid_idx on users(gitlab_id);

create table if not exists links (
    id serial primary key,
    link text not null,
    unique(link)
);

create index if not exists links_link_idx on links(link);

create table if not exists subs (
    id bigint primary key,
    link text unique not null,
    path text not null,
    type text not null,
    parent_id bigint,

    unique (id, parent_id)
);

create index if not exists subs_id_idx on subs(id);
create index if not exists subs_link_idx on subs(link);
create index if not exists subs_parentid_idx on subs(parent_id);

create table if not exists permissions (
    request_id text not null,
    user_id bigint,
    name text not null,
    username text not null,
    gitlab_id int not null,
    type text not null,
    link text not null,
    parent_link text,
    access_level int not null,
    created_at timestamp with time zone not null,
    expires_at timestamp without time zone not null,
    updated_at timestamp with time zone not null default current_timestamp,
    has_parent boolean not null default true,
    has_child boolean not null default false,

    unique (gitlab_id, link),
    foreign key(gitlab_id) references users(gitlab_id) on delete cascade on update cascade,
    foreign key(link) references subs(link) on delete cascade on update cascade
);

create index if not exists permissions_userid_idx on permissions(user_id);
create index if not exists permissions_gitlabid_idx on permissions(gitlab_id);
create index if not exists permissions_link_idx on permissions(link);

-- +migrate Down

drop index if exists permissions_link_idx;
drop index if exists permissions_gitlabid_idx;
drop index if exists permissions_userid_idx;

drop table if exists permissions;

drop index if exists subs_parentid_idx;
drop index if exists subs_link_idx;
drop index if exists subs_id_idx;

drop table if exists subs;

drop index if exists links_link_idx;

drop table if exists links;

drop index if exists users_gitlabid_idx;
drop index if exists users_username_idx;
drop index if exists users_id_idx;

drop table if exists users;
drop table if exists responses;
