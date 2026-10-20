-- Every table in the hub schema is now behind row-level security. The per-organization tables already were (rows of one
-- organization at a time). The rest are account-level or needed before an organization is known (accounts, sessions,
-- devices, key backups, data keys, organizations ...). They cannot be filtered by organization, so they are closed to any
-- connection that has not proven it is the Hub system: the application presents its guard token when it connects (the same
-- token that already guards writes). A leaked database password, an ad hoc console or a tool connecting as the application
-- role without the token now sees and changes nothing in them. The audit role (integrity check) and the schema owner
-- (migrations) keep their access.

create or replace function hub.is_app() returns boolean
  language sql stable security definer set search_path = hub, public
as $$
  select exists (
    select 1 from hub.guard g
    where g.id = 1 and g.token_hash = public.digest(coalesce(current_setting('hub.app_token', true), ''), 'sha256')
  )
$$;
revoke all on function hub.is_app() from public;
grant execute on function hub.is_app() to hub_app, hub_audit;

do $$
declare t text;
begin
  foreach t in array array['accounts','account_orgs','sessions','devices','key_backups','security_quiz','password_resets','login_blocks','org_keys','orgs','tones','blob_purge']
  loop
    execute format('alter table hub.%I enable row level security', t);
    execute format('alter table hub.%I force row level security', t);
    execute format('drop policy if exists hub_system on hub.%I', t);
    execute format($p$create policy hub_system on hub.%I for all using (hub.is_app() or current_user in ('hub_audit','hub_owner')) with check (hub.is_app() or current_user = 'hub_owner')$p$, t);
  end loop;
  -- the guard token and the migration list are never readable by the application at all: no policy means no rows
  alter table hub.guard enable row level security;
  alter table hub.migrations enable row level security;
end $$;
