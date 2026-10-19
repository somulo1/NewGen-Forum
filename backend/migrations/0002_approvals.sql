-- People can be removed from an organization without losing history.
alter table hub.people add column deleted_at timestamptz;
alter table hub.people drop constraint people_status_check;
alter table hub.people add constraint people_status_check check (status in ('active','inactive','blocked','deleted'));

-- ---------------------------------------------------------------------------------------------
-- Approvals engine. A flow says how many levels of approval an action needs and who approves at each level.
-- A request carries the exact change to apply once the last level approves (encrypted: it can contain personal data).
-- ---------------------------------------------------------------------------------------------
create table hub.approval_flows (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  key text not null, module text not null, action text not null,
  levels int not null check (levels between 0 and 3), approvers text[] not null,
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), unique (org_id, key)
);

create table hub.approval_requests (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  seq int not null, flow_key text not null, kind text not null,
  title_enc bytea not null, note_enc bytea, payload_enc bytea,
  requester uuid not null references hub.people(id),
  level int not null default 1, status text not null default 'Pending' check (status in ('Pending','Approved','Rejected','Cancelled')),
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(), decided_at timestamptz,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), unique (org_id, seq)
);
create index approval_requests_status_idx on hub.approval_requests (org_id, status, created_at desc);

create table hub.approval_steps (
  org_id uuid not null references hub.orgs(id), request_id uuid not null references hub.approval_requests(id),
  level int not null, approver uuid not null references hub.people(id),
  decision text not null check (decision in ('Approved','Rejected')), comment_enc bytea,
  created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, request_id, level)
);

-- Per-organization counters (AP-33, AP-34 ...) allocated inside the transaction that needs them.
create table hub.counters (
  org_id uuid not null references hub.orgs(id), name text not null, value bigint not null default 0, primary key (org_id, name)
);

select hub.protect('hub.approval_flows');
select hub.protect('hub.approval_requests');
select hub.protect('hub.approval_steps');
select hub.protect('hub.counters', true, 'guard');
