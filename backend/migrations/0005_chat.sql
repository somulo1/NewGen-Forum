-- Forum names and descriptions can be sensitive, so they are encrypted at rest with the organization's data key.
alter table hub.conversations drop column name, drop column about;
alter table hub.conversations add column name_enc bytea, add column about_enc bytea;
alter table hub.conversations add constraint conversations_privacy_check check (privacy in ('invite','open'));

-- Latest activity per conversation, kept apart from the conversation row so a busy forum does not rewrite (and re-seal) it
-- for every message.
create table hub.conversation_stats (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  last_message uuid, last_at timestamptz, message_count bigint not null default 0,
  primary key (org_id, conversation_id)
);
select hub.protect('hub.conversation_stats', true, 'guard');
create index conversation_stats_recent_idx on hub.conversation_stats (org_id, last_at desc);
