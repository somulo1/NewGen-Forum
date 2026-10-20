-- User groups: a named set of roles. Rights and widget access can be given to a group, and everyone whose role is in the
-- group gets them. A person still has one role; the group only adds to what the role allows.
create table hub.role_groups (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  key text not null, name text not null, description text not null default '',
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), unique (org_id, key)
);
create unique index role_groups_name_idx on hub.role_groups (org_id, lower(name));
create table hub.role_group_roles (
  org_id uuid not null references hub.orgs(id), group_id uuid not null references hub.role_groups(id), role_id uuid not null references hub.roles(id),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, group_id, role_id)
);
create index role_group_roles_role_idx on hub.role_group_roles (role_id);
create table hub.role_group_rights (
  org_id uuid not null references hub.orgs(id), group_id uuid not null references hub.role_groups(id),
  module text not null, right_name text not null, allowed boolean not null,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, group_id, module, right_name)
);
select hub.protect('hub.role_groups');
select hub.protect('hub.role_group_roles');
select hub.protect('hub.role_group_rights');

-- Does a role hold a right: its own setting, or a group it belongs to allowing it. (A personal override on a person still
-- beats both; the callers handle that.) Row-level security scopes it to the caller's organization.
create function hub.role_right(p_role uuid, p_module text, p_right text) returns boolean language sql stable as $$
  select coalesce((select rr.allowed from hub.role_rights rr where rr.role_id = p_role and rr.module = p_module and rr.right_name = p_right), false)
      or exists (select 1 from hub.role_group_roles gr join hub.role_group_rights g on g.group_id = gr.group_id
                 where gr.role_id = p_role and g.module = p_module and g.right_name = p_right and g.allowed)
$$;
