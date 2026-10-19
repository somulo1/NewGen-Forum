#!/usr/bin/env bash
# One-time bootstrap for a managed Postgres (Neon etc.) where you cannot use a local superuser socket.
# Same result as bootstrap_db.sh: creates hub_owner / hub_app / hub_audit / hub_dba inside an existing database.
#   HUB_ADMIN_URL   the provider's admin/owner connection string (use the DIRECT host, not the -pooler one)
#   HUB_*_PW        role passwords, as in .env
# Usage: set -a; source .env; set +a; HUB_ADMIN_URL='postgresql://...' bash scripts/bootstrap_remote.sh
set -euo pipefail
: "${HUB_ADMIN_URL:?set HUB_ADMIN_URL}"
: "${HUB_OWNER_PW:?}" "${HUB_APP_PW:?}" "${HUB_AUDIT_PW:?}" "${HUB_DBA_PW:?}"
psql "$HUB_ADMIN_URL" -v ON_ERROR_STOP=1 -v owner_pw="$HUB_OWNER_PW" -v app_pw="$HUB_APP_PW" -v audit_pw="$HUB_AUDIT_PW" -v dba_pw="$HUB_DBA_PW" <<'SQL'
select format('create role hub_owner login password %L nosuperuser nocreaterole nocreatedb', :'owner_pw')
  where not exists (select 1 from pg_roles where rolname='hub_owner') \gexec
select format('create role hub_app   login password %L nosuperuser nocreaterole nocreatedb nobypassrls', :'app_pw')
  where not exists (select 1 from pg_roles where rolname='hub_app') \gexec
select format('create role hub_audit login password %L nosuperuser nocreaterole nocreatedb nobypassrls', :'audit_pw')
  where not exists (select 1 from pg_roles where rolname='hub_audit') \gexec
select format('create role hub_dba   login password %L nosuperuser nocreaterole nocreatedb nobypassrls', :'dba_pw')
  where not exists (select 1 from pg_roles where rolname='hub_dba') \gexec
create extension if not exists pgcrypto;
select format('grant create, connect on database %I to hub_owner', current_database()) \gexec
select format('grant connect on database %I to hub_app, hub_audit, hub_dba', current_database()) \gexec
revoke all on schema public from public;
grant usage on schema public to hub_owner, hub_app, hub_audit, hub_dba;
SQL
echo "bootstrap ok"
