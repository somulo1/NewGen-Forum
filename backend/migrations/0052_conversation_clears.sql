-- "Delete" a chat or forum for me: my copy of its history up to this moment is gone and the conversation leaves my sidebar.
-- Other members keep theirs. When someone writes there again it comes back, holding only what was written after this time.
create table hub.conversation_clears (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  account_id uuid not null references hub.accounts(id), cleared_at timestamptz not null default now(),
  primary key (org_id, conversation_id, account_id)
);
select hub.protect('hub.conversation_clears', true, 'guard');
