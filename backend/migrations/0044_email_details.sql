-- What an email may say about a ticket or a communication. Their content is end-to-end encrypted, so the server cannot read it;
-- when the organization allows details in emails (security setting), the sender's app also gives the server a copy, kept here
-- encrypted with the organization's key, and only used to write emails.
create table hub.email_details (
  org_id uuid not null references hub.orgs(id), id uuid not null, kind text not null check (kind in ('ticket','comm')),
  data_enc bytea not null, created_at timestamptz not null default now(), primary key (org_id, id)
);
select hub.protect('hub.email_details', true, 'guard');
