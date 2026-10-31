-- A person's deleted call-log entries. Deleting is for good and for that person only: the entry never comes back in their
-- history, and the other people on the call keep theirs.
create table hub.call_log_deletes (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id), call_id uuid not null,
  deleted_at timestamptz not null default now(), primary key (org_id, person_id, call_id)
);
select hub.protect('hub.call_log_deletes', true, 'guard');
