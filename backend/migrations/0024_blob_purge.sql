-- Contents still to be erased from storage. A deleted file's bytes are erased right after the delete commits; if the file
-- server cannot be reached then, the reference waits here and the worker keeps trying until the bytes are gone.
create table hub.blob_purge (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id), ref text not null,
  attempts int not null default 0, last_error text, next_at timestamptz not null default now(),
  created_at timestamptz not null default now()
);
create index blob_purge_due_idx on hub.blob_purge (next_at);
select hub.protect('hub.blob_purge', false, 'guard');
