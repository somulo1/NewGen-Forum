-- Job posts join memos, announcements and newsletters. Everything sent in one go (to a wall and some forums) shares a batch, so
-- its interaction can be read together.
alter table hub.comms drop constraint if exists comms_type_check;
alter table hub.comms add constraint comms_type_check check (type in ('Memo','Announcement','Newsletter','Job post'));
alter table hub.comms add column if not exists batch uuid;
create index if not exists comms_batch_idx on hub.comms (org_id, batch) where batch is not null;
