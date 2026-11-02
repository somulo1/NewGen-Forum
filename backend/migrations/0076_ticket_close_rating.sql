-- Closing a ticket: when it was closed, the stars it was given (1 to 5) and whether the system closed it (30 minutes after it was
-- resolved and nobody closed it). A comment written by the system has no author.
alter table hub.tickets add column closed_at timestamptz, add column rating smallint check (rating between 1 and 5), add column auto_closed boolean not null default false;
alter table hub.ticket_comments alter column author drop not null;
create index if not exists tickets_resolved_idx on hub.tickets (org_id, solved_at) where step = 3;
