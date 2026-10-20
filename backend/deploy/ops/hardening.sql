-- Run once as a superuser on the production database after scripts/bootstrap_db.sh has made the roles.
-- The application never connects as owner or superuser; this checks that and adds a backup role.
create role hub_backup login password :'backup_pw' nosuperuser nocreatedb nocreaterole;
grant pg_read_all_data to hub_backup;                       -- read only, for pg_dump
revoke create on schema public from public;
alter role hub_app set statement_timeout = '30s';           -- a runaway query cannot hold a connection for long
alter role hub_app set idle_in_transaction_session_timeout = '15s';
alter role hub_app connection limit 200;                    -- keep below max_connections minus other roles
-- Proof that the app role cannot do what an attacker would want (each must be false):
select rolsuper, rolcreatedb, rolcreaterole, rolbypassrls from pg_roles where rolname = 'hub_app';
