create table if not exists clone_history
(
    id           serial    not null
        constraint clone_history_pkey primary key,
    created_at   timestamp not null default CURRENT_TIMESTAMP,
    name         varchar   not null default '',
    project_id   integer   not null default 0,
    cloning_path varchar   not null default '',
    release_tag  varchar   not null default '',
    structure    jsonb     not null default '{}'
);
