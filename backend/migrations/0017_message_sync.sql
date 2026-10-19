-- Catch-up reads: "what changed in this conversation since T" (new messages by when the server stored them, removals by
-- when they happened), so an open chat fetches only what is new instead of re-reading its whole history.
create index if not exists messages_conv_created_idx on hub.messages (org_id, conversation_id, created_at, id);
create index if not exists messages_conv_deleted_idx on hub.messages (org_id, conversation_id, deleted_at) where deleted_at is not null;
