-- Tickets are no longer end-to-end encrypted at all: their attachments are plain files the server keeps for the people on the
-- ticket, so raising, opening and handing on a ticket needs no keys. Attachments added before this stay in the ticket's
-- conversation, as they were.
create table hub.ticket_files (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), ticket_id uuid not null references hub.conversations(id),
  uploader uuid not null references hub.people(id), name text not null, type text not null default '', size bigint not null,
  blob_key text not null, created_at timestamptz not null default now(),
  primary key (org_id, id)
);
create index ticket_files_idx on hub.ticket_files (org_id, ticket_id, created_at);
select hub.protect('hub.ticket_files', true, 'guard');
