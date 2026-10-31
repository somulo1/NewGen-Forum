-- Which uploaded files a message carries. The server cannot read a sealed message, so it is told at send time. This is what lets a
-- deleted post, or a moment after its 24 hours, take its pictures and videos with it: the files are erased from storage, not just hidden.
create table hub.message_files (
  org_id uuid not null references hub.orgs(id), message_id uuid not null, file_id uuid not null, primary key (org_id, file_id)
);
create index message_files_msg_idx on hub.message_files (org_id, message_id);
select hub.protect('hub.message_files', true, 'guard');
