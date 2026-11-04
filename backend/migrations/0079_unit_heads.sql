-- The role that heads a department, sector, branch or unit. Everyone who holds that role and is placed in the unit heads it, with
-- whatever the role allows. It is its own table (not a column on org_units), so the seals of existing units stay valid. role_id has no
-- foreign key on purpose: deleting a role must never be blocked by it, and a reference to a missing role simply shows nothing.
create table hub.unit_heads (
  org_id uuid not null references hub.orgs(id), unit_id uuid not null, role_id uuid not null,
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, unit_id)
);
select hub.protect('hub.unit_heads');
