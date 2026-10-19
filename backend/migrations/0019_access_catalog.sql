-- The rights catalog is data: each organization has its own list of modules (pages) and the rights inside them. Roles are created
-- by the organization and given rights from this list. Rights the server enforces are marked builtin (their label can change,
-- they cannot be removed); everything else can be added or removed.
create table hub.modules (
  org_id uuid not null references hub.orgs(id), key text not null, label text not null, tint text not null default '#9DB0A6',
  sort int not null default 0, builtin boolean not null default false,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, key)
);
create table hub.rights (
  org_id uuid not null references hub.orgs(id), module text not null, name text not null, label text not null default '',
  kind text not null default 'other' check (kind in ('view','create','edit','delete','export','approve','manage','other')),
  sort int not null default 0, builtin boolean not null default false,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, module, name),
  foreign key (org_id, module) references hub.modules (org_id, key) on delete cascade
);
select hub.protect('hub.modules');
select hub.protect('hub.rights');

-- The role new people get when none is chosen (at most one per organization).
create table hub.default_role (
  org_id uuid not null references hub.orgs(id) primary key, role_id uuid not null references hub.roles(id),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
select hub.protect('hub.default_role');

-- Who approves a level is decided by a right ("Approve level N" in the flow's module), not by a name, so the approvers list
-- on a flow is no longer stored.
alter table hub.approval_flows alter column approvers set default '{}';
