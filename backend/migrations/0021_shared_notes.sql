-- A note shared with other people lives in its own small conversation (the owner, plus people who can read or edit it).
alter table hub.conversations drop constraint conversations_kind_check;
alter table hub.conversations add constraint conversations_kind_check check (kind in ('dm','forum','wall','note','ticket','call','project','event','folder','sharednote'));
