-- Who was working on a ticket when its stay with a place ended, so each person's resolutions and escalations count for them.
alter table hub.ticket_hops add column handler uuid;
