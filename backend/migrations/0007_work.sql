-- Projects: the plain facts the server needs (status, dates, owner). Names, goals, tasks and discussion are encrypted.
create table hub.projects (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  code text not null, color text not null default '#7CC4FF', owner uuid not null references hub.people(id),
  status text not null default 'Pending approval' check (status in ('Pending approval','On track','At risk','Off track','Delayed','Done','Cancelled')),
  start_on date, due_on date, budget_pct int check (budget_pct between 0 and 100), sync text not null default '', links text[] not null default '{}',
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, conversation_id), unique (org_id, code)
);
select hub.protect('hub.projects');

-- Tickets. Routing, timing and status are plain (the server must act on them); the subject, description and thread are
-- end-to-end encrypted in the ticket's conversation.
create table hub.ticket_settings (
  org_id uuid primary key references hub.orgs(id), config jsonb not null default '{}',
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
create table hub.ticket_categories (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(),
  unit_type text not null check (unit_type in ('desk','department','unit','branch','sector')), unit_ref uuid,
  name text not null, tat_hours int not null check (tat_hours between 1 and 24*90), priority text not null default 'Medium' check (priority in ('Low','Medium','High')),
  members uuid[] not null default '{}',
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, id)
);
create index ticket_categories_unit_idx on hub.ticket_categories (org_id, unit_type, unit_ref);
create table hub.tickets (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  ref text not null, category text not null, priority text not null check (priority in ('Low','Medium','High')),
  step smallint not null default 0 check (step between 0 and 4), hold boolean not null default false,
  raised_by uuid not null references hub.people(id), handler uuid references hub.people(id), handler_since timestamptz,
  to_type text not null default 'desk', to_ref uuid, tat_hours int not null, due_at timestamptz, solved_at timestamptz,
  escalated boolean not null default false, late boolean not null default false, sync text not null default '',
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, conversation_id), unique (org_id, ref)
);
create index tickets_open_idx on hub.tickets (org_id, step, due_at) where step < 3;
create index tickets_handler_idx on hub.tickets (org_id, handler) where step < 3;
create table hub.ticket_assignees (
  org_id uuid not null references hub.orgs(id), ticket_id uuid not null references hub.conversations(id), person_id uuid not null references hub.people(id),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, ticket_id, person_id)
);
-- What happened to a ticket, written by the server as it acts. Append-only.
create table hub.ticket_events (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), ticket_id uuid not null references hub.conversations(id),
  at timestamptz not null default now(), actor uuid, actor_name text not null, kind text not null, text text not null,
  primary key (org_id, id)
);
create index ticket_events_idx on hub.ticket_events (org_id, ticket_id, at);
select hub.protect('hub.ticket_settings');
select hub.protect('hub.ticket_categories');
select hub.protect('hub.tickets');
select hub.protect('hub.ticket_assignees');
select hub.protect('hub.ticket_events', true, 'readonly');
