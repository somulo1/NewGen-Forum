-- A wall attachment belongs to the post or moment that carries it, so deleting that post erases the file too.
alter table hub.plain_files add column message_id uuid;
create index plain_files_message_idx on hub.plain_files (org_id, message_id) where message_id is not null;
