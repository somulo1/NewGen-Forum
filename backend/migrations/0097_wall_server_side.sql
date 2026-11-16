-- The community wall is fully server-side: posts, comments and moments are read and kept by the server (encrypted at rest with
-- the organization's key, see chat/plain.go), likes and votes are rows the server counts, and attachments are plain files kept
-- for the organization. Running totals mean a post with a hundred thousand likes still loads in one small read.
create table hub.wall_reactions (
  org_id uuid not null references hub.orgs(id), item_id uuid not null, person_id uuid not null references hub.people(id),
  kind text not null check (kind in ('like','dislike','love')), at timestamptz not null default now(),
  primary key (org_id, item_id, person_id)
);
create index wall_reactions_item_idx on hub.wall_reactions (org_id, item_id, kind, at);
select hub.protect('hub.wall_reactions', true, 'guard');

create table hub.wall_votes (
  org_id uuid not null references hub.orgs(id), post_id uuid not null, person_id uuid not null references hub.people(id),
  option text not null, at timestamptz not null default now(), primary key (org_id, post_id, person_id)
);
select hub.protect('hub.wall_votes', true, 'guard');

-- per item: 'like', 'dislike', 'love', 'vote:<option>'
create table hub.wall_counts (
  org_id uuid not null references hub.orgs(id), item_id uuid not null, key text not null, n int not null default 0,
  primary key (org_id, item_id, key)
);
select hub.protect('hub.wall_counts', true, 'guard');

-- attachments of server-readable conversations (the wall): plain files, readable by the conversation's members
create table hub.plain_files (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), conversation_id uuid not null references hub.conversations(id),
  uploader uuid not null references hub.people(id), name text not null, type text not null default '', size bigint not null,
  blob_key text not null, created_at timestamptz not null default now(), primary key (org_id, id)
);
create index plain_files_conv_idx on hub.plain_files (org_id, conversation_id, created_at);
select hub.protect('hub.plain_files', true, 'guard');
