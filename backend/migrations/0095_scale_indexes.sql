-- The ticket sweep looks up "already told this ticket is nearing its deadline" by ticket reference, for a whole batch at once.
create index if not exists notifications_ticket_near_idx on hub.notifications (org_id, ref) where kind = 'ticket-near';
-- Insights and reminders look up what one person has read across communications.
create index if not exists comm_reads_person_idx on hub.comm_reads (org_id, person_id);
