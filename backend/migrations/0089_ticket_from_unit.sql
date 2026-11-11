-- What a department, unit, branch or sector sends from its dashboard (a communication, or a business request) is from that place;
-- what a person raises from the ticket form is from them. Left empty for what people raise themselves.
alter table hub.tickets add column from_type text, add column from_ref uuid;
create index tickets_from_idx on hub.tickets (org_id, from_ref) where from_ref is not null;
