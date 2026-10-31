-- When anything last changed in a conversation (a message, an edit, a like or a removal). A device that was offline compares this
-- with what it holds and fetches only the conversations that changed, instead of asking about every one of them.
alter table hub.conversation_stats add column if not exists changed_at timestamptz not null default now();
create index if not exists conversation_stats_changed_idx on hub.conversation_stats (org_id, changed_at desc);
