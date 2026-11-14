-- A department (or any place) may choose to receive tickets only by escalation, and only from one chosen place. It then no longer
-- appears as somewhere to send a ticket or business request, and only the members of that chosen place see it (and its
-- categories) when they escalate.
create table hub.ticket_intake (
  org_id uuid not null references hub.orgs(id), unit_id uuid not null references hub.org_units(id) on delete cascade,
  only_from uuid not null references hub.org_units(id) on delete cascade, updated_at timestamptz not null default now(),
  primary key (org_id, unit_id), check (unit_id <> only_from)
);
select hub.protect('hub.ticket_intake', true, 'guard');
