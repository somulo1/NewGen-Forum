-- Each place a ticket or business request stays with: when it arrived, when it left, how (resolved or escalated) and whether that
-- was within that place's own turnaround. A place is judged only on its own stay, so an escalation never carries one place's
-- clock into the next. Older tickets without rows are read as one stay with the place they are at.
create table hub.ticket_hops (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), ticket_id uuid not null,
  unit_type text not null, unit_ref uuid, category text not null, tat_hours int not null,
  received_at timestamptz not null, left_at timestamptz, outcome text check (outcome in ('resolved','escalated')),
  due_at timestamptz, on_time boolean,
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, id)
);
create index ticket_hops_ticket_idx on hub.ticket_hops (org_id, ticket_id, received_at);
create index ticket_hops_unit_idx on hub.ticket_hops (org_id, unit_ref, received_at);
select hub.protect('hub.ticket_hops');
