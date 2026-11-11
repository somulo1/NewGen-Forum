-- A business request is raised from the same form as a ticket and to a department only. It is a ticket record marked as a request
-- (left empty for ordinary tickets, so the seals of existing tickets stay valid).
alter table hub.tickets add column kind text check (kind is null or kind = 'request');
create index tickets_kind_idx on hub.tickets (org_id, to_ref) where kind = 'request';
