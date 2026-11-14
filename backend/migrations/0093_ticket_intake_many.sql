-- A place may take escalations from several chosen places, not just one: one row per chosen place.
alter table hub.ticket_intake drop constraint ticket_intake_pkey;
alter table hub.ticket_intake add primary key (org_id, unit_id, only_from);
