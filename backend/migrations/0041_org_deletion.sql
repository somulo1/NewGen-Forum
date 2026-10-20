-- An organization that asked for its data to be deleted: everything is erased once this time passes (30 days after the
-- request, so a mistake can still be cancelled).
alter table hub.orgs add column delete_after timestamptz;

-- Append-only tables refuse deletes, except while the system erases one whole organization that reached its deletion
-- date (the eraser sets hub.purge_org for its own transaction only; writes still need the system's guard token).
create or replace function hub.deny_mutation() returns trigger language plpgsql as $$
begin
  if tg_op = 'DELETE' and tg_level = 'ROW' and coalesce(current_setting('hub.purge_org', true), '') <> ''
     and (to_jsonb(old) ->> 'org_id') = current_setting('hub.purge_org', true) then
    return old;
  end if;
  raise exception '%.% is append-only', tg_table_schema, tg_table_name using errcode = '42501';
end $$;
