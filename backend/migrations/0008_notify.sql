alter table hub.tickets add column hold_since timestamptz;

-- Notifications a person sees in the bell. They only name things the server already knows (a ticket reference, who handed it
-- over), never message content. Guarded rather than sealed: they are read/unread toggled constantly.
create table hub.notifications (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), person_id uuid not null references hub.people(id),
  kind text not null, text text not null, color text not null default '#7CC4FF', ref text not null default '',
  created_at timestamptz not null default now(), read_at timestamptz,
  primary key (org_id, id)
);
create index notifications_person_idx on hub.notifications (org_id, person_id, created_at desc);
select hub.protect('hub.notifications', true, 'guard');
