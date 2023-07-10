create table if not exists organization
(
    id          serial    not null
        constraint organization_pkey primary key,
    created_at  timestamp not null default CURRENT_TIMESTAMP,
    updated_at  timestamp not null default CURRENT_TIMESTAMP,
    last_update timestamp not null default CURRENT_TIMESTAMP,
    name        varchar   not null default '',
    github_url  varchar   not null default '',
    gitlab_url  varchar   not null default '',
    local_path  varchar   not null default ''
);


create table if not exists project
(
    id                        serial    not null
        constraint projects_pkey primary key,
    created_at                timestamp not null default CURRENT_TIMESTAMP,
    updated_at                timestamp not null default CURRENT_TIMESTAMP,
    pushed_at                 timestamp,

    type                      integer   not null default 0,
    organization_id           integer   not null default 0,
    name                      varchar   not null default '',
    local_path                varchar   not null default '',

    specification_url                varchar   not null default '',
    specification_last_commit_name   varchar   not null default '',
    specification_last_commit_time   timestamp not null default CURRENT_TIMESTAMP,
    specification_last_commit_author varchar   not null default '',
    specification_release_tag        varchar   not null default '',

    realisation_url                varchar   not null default '',
    realisation_last_commit_name   varchar   not null default '',
    realisation_last_commit_time   timestamp not null default CURRENT_TIMESTAMP,
    realisation_last_commit_author varchar   not null default '',
    realisation_release_tag        varchar   not null default '',

    last_structure            jsonb     not null default '{}'
);
