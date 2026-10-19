-- Memos, announcements and newsletters. The content waits here, sealed, until its approvals are done; then it is posted into
-- the audience's encrypted conversation as an ordinary message and this row remembers which one.
create table hub.comms (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(),
  ref text not null, type text not null check (type in ('Memo','Announcement','Newsletter')),
  status text not null default 'Pending approval' check (status in ('Pending approval','Published','Rejected','Stale','Cancelled')),
  by_person uuid not null references hub.people(id), conversation_id uuid not null references hub.conversations(id),
  message_id uuid not null, epoch int not null, device uuid not null, envelope bytea not null,
  created_at timestamptz not null default now(), published_at timestamptz,
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id), unique (org_id, ref)
);
create index comms_recent_idx on hub.comms (org_id, created_at desc);
create table hub.comm_reads (
  org_id uuid not null references hub.orgs(id), comm_id uuid not null, person_id uuid not null references hub.people(id), at timestamptz not null default now(),
  primary key (org_id, comm_id, person_id)
);
create table hub.comm_rules (
  org_id uuid not null references hub.orgs(id), level text not null check (level in ('departments','branches','sectors','units')),
  type text not null check (type in ('Memo','Announcement')), config jsonb not null,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, level, type)
);

-- Performance awards.
create table hub.awards (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(),
  ref text not null, type text not null, people uuid[] not null, citation_enc bytea not null,
  on_date date not null default current_date, status text not null default 'Pending approval' check (status in ('Pending approval','Approved','Rejected')),
  nominated_by uuid not null references hub.people(id),
  created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id), unique (org_id, ref)
);
select hub.protect('hub.comms');
select hub.protect('hub.comm_reads', true, 'guard');
select hub.protect('hub.comm_rules');
select hub.protect('hub.awards');
