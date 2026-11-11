-- Acknowledging is a click that says "received and noted", apart from just opening it. For organization-wide communications
-- (memos, announcements, newsletters) its insights show who did; for business requests and communications sent between places,
-- the people of the receiving place do it on the item.
create table hub.comm_acks (
  org_id uuid not null references hub.orgs(id), comm_id uuid not null, person_id uuid not null references hub.people(id), at timestamptz not null default now(),
  primary key (org_id, comm_id, person_id)
);
select hub.protect('hub.comm_acks', true, 'guard');
create table hub.ticket_acks (
  org_id uuid not null references hub.orgs(id), ticket_id uuid not null, person_id uuid not null references hub.people(id), at timestamptz not null default now(),
  primary key (org_id, ticket_id, person_id)
);
select hub.protect('hub.ticket_acks', true, 'guard');
