#!/usr/bin/env bash
# One-time database bootstrap, run by a superuser. Creates the roles and the database.
#   hub_owner  owns the schema; used only by cmd/migrate. Cannot be used by the running service.
#   hub_app    the running service: DML on tables, no DDL, bound by row level security and the app guard.
#   hub_audit  read-only access for auditors and the integrity scanner.
#   hub_dba    for humans: can inspect structure and run maintenance, but has NO insert/update/delete on data.
# Passwords come from the environment in real deployments; the defaults below are for local development only.
set -euo pipefail
PSQL=${PSQL:-psql -h /tmp -p 5433 -U hubadmin}
OWNER_PW=${HUB_OWNER_PW:-hub_owner_dev_pw}; APP_PW=${HUB_APP_PW:-hub_app_dev_pw}
AUDIT_PW=${HUB_AUDIT_PW:-hub_audit_dev_pw}; DBA_PW=${HUB_DBA_PW:-hub_dba_dev_pw}
$PSQL -d postgres -v ON_ERROR_STOP=1 <<SQL
do \$\$ begin
  if not exists (select 1 from pg_roles where rolname='hub_owner') then create role hub_owner login password '$OWNER_PW' nosuperuser nocreaterole nocreatedb; end if;
  if not exists (select 1 from pg_roles where rolname='hub_app')   then create role hub_app   login password '$APP_PW'   nosuperuser nocreaterole nocreatedb nobypassrls; end if;
  if not exists (select 1 from pg_roles where rolname='hub_audit') then create role hub_audit login password '$AUDIT_PW' nosuperuser nocreaterole nocreatedb nobypassrls; end if;
  if not exists (select 1 from pg_roles where rolname='hub_dba')   then create role hub_dba   login password '$DBA_PW'   nosuperuser nocreaterole nocreatedb nobypassrls; end if;
end \$\$;
SQL
$PSQL -d postgres -tc "select 1 from pg_database where datname='hub'" | grep -q 1 || $PSQL -d postgres -c "create database hub owner hub_owner encoding 'UTF8'"
$PSQL -d hub -v ON_ERROR_STOP=1 <<SQL
revoke all on schema public from public;
grant usage on schema public to hub_owner;
create extension if not exists pgcrypto;
grant connect on database hub to hub_app, hub_audit, hub_dba;
SQL
echo "bootstrap ok"
