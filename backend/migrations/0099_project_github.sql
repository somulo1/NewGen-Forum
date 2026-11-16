-- A project may be linked to one GitHub repository (optional). GitHub sends its events to a webhook made for that project; Hub
-- matches branches and pull requests to tasks (by their stored branch name, or by the task key in a branch name) and either
-- suggests the next status for someone to accept, or moves the task itself when the project asks for that.
create table hub.project_github (
  org_id uuid not null references hub.orgs(id), project_id uuid not null references hub.conversations(id),
  repo text not null, hook_id uuid not null unique, secret_enc bytea not null, token_enc bytea,
  auto boolean not null default false, created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  primary key (org_id, project_id)
);
select hub.protect('hub.project_github', true, 'guard');

-- where a webhook delivery belongs (read before the organization is known, so not limited to one organization)
create table hub.github_hooks (
  hook_id uuid primary key, org_id uuid not null references hub.orgs(id), project_id uuid not null
);
select hub.protect('hub.github_hooks', false, 'guard');

-- what GitHub reported, per project and (when matched) task, with the status it suggests and what became of the suggestion
create table hub.project_git_events (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), project_id uuid not null, task_id uuid,
  delivery text, kind text not null, branch text not null default '', title text not null default '', url text not null default '',
  actor text not null default '', at timestamptz not null default now(),
  suggest text, state text not null default 'none' check (state in ('none','pending','accepted','dismissed','auto')),
  decided_by uuid, decided_at timestamptz,
  primary key (org_id, id)
);
create index project_git_events_idx on hub.project_git_events (org_id, project_id, at desc);
create index project_git_events_task_idx on hub.project_git_events (org_id, task_id, at desc) where task_id is not null;
create unique index project_git_events_delivery_idx on hub.project_git_events (org_id, project_id, delivery, coalesce(task_id, '00000000-0000-0000-0000-000000000000'::uuid)) where delivery is not null;
select hub.protect('hub.project_git_events', true, 'guard');
