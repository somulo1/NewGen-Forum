-- Departments, units, branches and sectors send each other communications from their dashboards (two-way, separate from the
-- organization-wide memos and announcements). Like business requests, each is a ticket record marked by its kind.
alter table hub.tickets drop constraint if exists tickets_kind_check;
alter table hub.tickets add constraint tickets_kind_check check (kind is null or kind in ('request', 'communication'));
