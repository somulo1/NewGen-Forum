-- A ticket's subject, description and comments are no longer end-to-end encrypted: the server reads them (search, reports,
-- emails) and keeps them encrypted at rest with the organization's data key, each value bound to its row. Attachments stay
-- end-to-end encrypted in the ticket's conversation. Tickets raised before this keep their text sealed in the conversation.
create table hub.ticket_text (
  org_id uuid not null references hub.orgs(id), ticket_id uuid not null references hub.conversations(id),
  title_enc bytea not null, description_enc bytea, created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, ticket_id)
);
select hub.protect('hub.ticket_text');

create table hub.ticket_comments (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), ticket_id uuid not null references hub.conversations(id),
  author uuid not null references hub.people(id), body_enc bytea not null, created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(), ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id)
);
create index ticket_comments_idx on hub.ticket_comments (org_id, ticket_id, created_at);
select hub.protect('hub.ticket_comments');
