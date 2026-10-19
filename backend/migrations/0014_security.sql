-- Security alerts raised by rules, and the notes people write as they work them. A note is required to change an alert's
-- status, so the reason is always on record.
create table hub.security_rules (
  org_id uuid not null references hub.orgs(id), id text not null check (id ~ '^r[0-9]+$'), enabled boolean not null default true,
  severity text not null default 'Medium' check (severity in ('High','Medium','Low')),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, id)
);
create table hub.alerts (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), ref text not null, rule text not null,
  severity text not null, title text not null, detail text not null default '',
  status text not null default 'Open' check (status in ('Open','Investigating','Resolved','False alarm')),
  created_at timestamptz not null default now(), closed_at timestamptz, ledger_seq bigint,
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id), unique (org_id, ref)
);
create index alerts_open_idx on hub.alerts (org_id, status, created_at desc);
create table hub.alert_notes (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), alert_id uuid not null,
  status text not null, comment_enc bytea not null, person uuid not null references hub.people(id), at timestamptz not null default now(),
  primary key (org_id, id)
);
create index alert_notes_idx on hub.alert_notes (org_id, alert_id, at);
select hub.protect('hub.security_rules');
select hub.protect('hub.alerts');
select hub.protect('hub.alert_notes', true, 'readonly');
