-- The performance threshold (percent of what a sector, branch, unit or department receives that should be handled on time).
create table hub.unit_targets (
  org_id uuid not null references hub.orgs(id), unit_id uuid not null references hub.org_units(id),
  target_pct int not null check (target_pct between 50 and 100), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, unit_id)
);
select hub.protect('hub.unit_targets');
