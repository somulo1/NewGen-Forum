-- A project's plan is no longer end-to-end encrypted: tasks, milestones, sprints, risks, activity and documents are plain rows the
-- server can track. Its discussion stays in the encrypted conversation. Older encrypted plans are moved over by a member's browser
-- (only a device can read them): plan_imported_at stays empty until then.
alter table hub.projects
  add column method text not null default 'agile' check (method in ('agile','waterfall','hybrid')),
  add column workflow jsonb not null default '[{"key":"todo","name":"To do","cat":0},{"key":"doing","name":"In progress","cat":1},{"key":"review","name":"Review","cat":2},{"key":"done","name":"Done","cat":3}]',
  add column plan_imported_at timestamptz;
-- a plan with nothing in it has nothing to move (read as owner: this file runs with triggers guarding only the app role)
update hub.projects p set plan_imported_at = now()
  where not exists (select 1 from hub.records r where r.container_id = p.conversation_id and r.collection in ('task','milestone','sprint','risk','activity') and r.deleted_at is null);

create table hub.project_items (
  org_id uuid not null references hub.orgs(id), id uuid not null, project_id uuid not null references hub.conversations(id),
  collection text not null check (collection in ('task','milestone','sprint','risk','activity','page')),
  ref int, data jsonb not null, created_by uuid references hub.people(id), created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(), ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id)
);
create index project_items_idx on hub.project_items (org_id, project_id, collection);
create unique index project_items_ref_idx on hub.project_items (org_id, project_id, ref) where ref is not null;
select hub.protect('hub.project_items');

-- Every change to every field: who, when, from what, to what. Added to only; a document's versions are read from here.
create table hub.project_item_history (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), item_id uuid not null, project_id uuid not null,
  person_id uuid, who text not null, at timestamptz not null default now(), field text not null, old_value text, new_value text,
  primary key (org_id, id)
);
create index project_item_history_idx on hub.project_item_history (org_id, project_id, item_id, at);
select hub.protect('hub.project_item_history', true, 'guard');
