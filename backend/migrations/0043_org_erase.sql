-- Erasing an organization's append-only records (logs, ledger, key history). The app role has no delete right on these
-- tables, by design. This function, owned by the database owner, is the one way in, and it refuses unless the
-- organization's 30-day deletion wait is over.
create function hub.erase_append_only(p_org uuid) returns void language plpgsql security definer set search_path = hub, public, pg_temp as $$
begin
  if not exists (select 1 from hub.orgs where id = p_org and delete_after is not null and delete_after <= now()) then
    raise exception 'organization % is not due for deletion', p_org using errcode = '42501';
  end if;
  perform set_config('hub.purge_org', p_org::text, true);
  delete from hub.ledger_checkpoints where org_id = p_org;
  delete from hub.ledger where org_id = p_org;
  delete from hub.audit_logs where org_id = p_org;
  delete from hub.org_keys where org_id = p_org;
end $$;
revoke all on function hub.erase_append_only(uuid) from public;
