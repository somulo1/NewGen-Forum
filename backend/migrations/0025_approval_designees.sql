-- The person chosen to decide one level of an approval flow. A level with no row is decided by anyone holding the right
-- "Approve level N"; a level with a row is decided only by that person (the owner role can always stand in).
create table hub.approval_designees (
  org_id uuid not null references hub.orgs(id), flow_key text not null, level int not null check (level between 1 and 3),
  person_id uuid not null references hub.people(id), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, flow_key, level)
);
select hub.protect('hub.approval_designees');
