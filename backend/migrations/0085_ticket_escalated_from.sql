-- The units a ticket was escalated away from, so each one still sees it as escalated on its dashboard after it moved on.
alter table hub.tickets add column escalated_from uuid[];
