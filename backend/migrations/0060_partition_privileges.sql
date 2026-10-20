-- Row-level security is declared on the parent tables (messages, records), and PostgreSQL does not apply a parent's policy to a
-- query that names a partition directly. The default privileges gave the application and audit roles access to every
-- partition, so a flaw that let someone run `select * from hub.messages_p8` would have read every organization's rows.
-- The roles keep full access through the parent (PostgreSQL checks privileges on the parent only); direct access to the
-- partitions is taken away from them.
do $$
declare t record;
begin
  for t in
    select n.nspname, c.relname from pg_class c join pg_namespace n on n.oid = c.relnamespace
    where n.nspname = 'hub' and c.relispartition and c.relkind = 'r'
  loop
    execute format('revoke all on %I.%I from hub_app, hub_audit, hub_dba', t.nspname, t.relname);
  end loop;
end $$;
