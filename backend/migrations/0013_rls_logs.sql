-- The audit log and the ledger are per organization like everything else: one organization can never read another's.
do $$ declare t text; begin
  foreach t in array array['hub.audit_logs','hub.ledger','hub.ledger_heads','hub.ledger_checkpoints'] loop
    execute format('alter table %s enable row level security', t);
    execute format('alter table %s force row level security', t);
    execute format('create policy tenant on %s using (org_id = hub.current_org()) with check (org_id = hub.current_org())', t);
  end loop;
end $$;
