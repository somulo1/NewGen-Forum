-- 0001 core: tenancy, identity, key directory, E2EE conversation storage, tamper-evident ledger.
-- Never edit an applied migration: the runner stores each file's SHA-256 and refuses to start if one changed.

create schema if not exists hub;
create extension if not exists pgcrypto;

-- ---------------------------------------------------------------------------------------------
-- Guards. Data may only change through the running system:
--  * hub_app has DML but no DDL; humans use roles with no DML at all (see scripts/bootstrap_db.sh)
--  * every write must come from a connection that proves it is the app (guard token set at connect)
--  * mutable rows carry a row MAC keyed with a secret that never lives in the database
--  * ledger and key tables are append-only (updates and deletes are refused)
-- ---------------------------------------------------------------------------------------------
create table hub.guard (id smallint primary key check (id = 1), token_hash bytea not null);

create function hub.current_org() returns uuid language sql stable as
$$ select nullif(current_setting('hub.org_id', true), '')::uuid $$;

create function hub.assert_app() returns trigger language plpgsql security definer set search_path = hub, public, pg_temp as $$
declare g bytea;
begin
  select token_hash into g from hub.guard where id = 1;
  if g is null or public.digest(coalesce(current_setting('hub.app_token', true), ''), 'sha256') <> g then
    raise exception 'direct data changes are not allowed on %.%: data may only be written by the Hub system', tg_table_schema, tg_table_name using errcode = '42501';
  end if;
  if tg_op = 'DELETE' then return old; end if;
  return new;
end $$;

create function hub.deny_mutation() returns trigger language plpgsql as $$
begin
  raise exception '%.% is append-only', tg_table_schema, tg_table_name using errcode = '42501';
end $$;

-- Row integrity. Every mutable row carries a MAC (HMAC-SHA256 under a key that never lives in the database) over its
-- canonical content, computed by the system after each write. The database enforces the protocol:
--   insert  -> row arrives marked "pending", the system reads back its canonical form, then seals it with the real MAC
--   update  -> version +1 and "pending", then sealed again
-- A script that edits a row but cannot produce the MAC is stopped here; one that bypasses the triggers leaves a row whose
-- MAC no longer matches, which the integrity scanner reports.
create function hub.pending_mac() returns bytea language sql immutable as
$$ select '\xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff'::bytea $$;

create function hub.canon(r anyelement) returns text language sql stable set timezone = 'UTC' as
$$ select (to_jsonb(r) - 'mac' - 'ver' - 'updated_at')::text $$;

create function hub.require_mac() returns trigger language plpgsql as $$
begin
  if tg_op = 'INSERT' then
    if new.mac is distinct from hub.pending_mac() or new.ver <> 1 then
      raise exception 'new rows on %.% must start pending (version 1) and be sealed by the system', tg_table_schema, tg_table_name using errcode = '23514';
    end if;
  elsif old.mac = hub.pending_mac() then            -- sealing step: same version, real MAC
    if new.ver <> old.ver or new.mac = hub.pending_mac() or length(new.mac) <> 32 then
      raise exception 'invalid seal on %.%', tg_table_schema, tg_table_name using errcode = '23514';
    end if;
  else                                               -- editing step: next version, pending again
    if new.ver <> old.ver + 1 or new.mac is distinct from hub.pending_mac() then
      raise exception 'edits to %.% must advance the version and be sealed by the system', tg_table_schema, tg_table_name using errcode = '23514';
    end if;
  end if;
  return new;
end $$;

-- Helper that wires the standard guards onto a table.
--   mode 'mac'      mutable row, needs a row MAC and a version bump on every write
--   mode 'guard'    mutable but only through the app (no MAC): small operational rows
--   mode 'readonly' append-only: no updates, no deletes
create function hub.protect(t regclass, tenant boolean default true, mode text default 'mac') returns void language plpgsql as $$
declare n text := replace(t::text, '.', '_');
begin
  execute format('create trigger %I before insert or update or delete on %s for each row execute function hub.assert_app()', n || '_app', t);
  execute format('create trigger %I before truncate on %s for each statement execute function hub.deny_mutation()', n || '_notrunc', t);
  if mode = 'mac' then
    execute format('create trigger %I before insert or update on %s for each row execute function hub.require_mac()', n || '_mac', t);
  elsif mode = 'readonly' then
    execute format('create trigger %I before update or delete on %s for each row execute function hub.deny_mutation()', n || '_ro', t);
  end if;
  if tenant then
    execute format('alter table %s enable row level security', t);
    execute format('alter table %s force row level security', t);
    execute format('create policy tenant on %s using (org_id = hub.current_org()) with check (org_id = hub.current_org())', t);
  end if;
end $$;

-- ---------------------------------------------------------------------------------------------
-- Organizations and their data keys
-- ---------------------------------------------------------------------------------------------
create table hub.orgs (
  id uuid primary key default uuidv7(),
  slug text not null unique check (slug ~ '^[a-z0-9][a-z0-9-]{1,40}$'),
  name text not null,
  plan text not null default 'business',
  status text not null default 'active' check (status in ('active','suspended','closing','deleted')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);

-- Wrapped per-organization data keys (the KEK that wraps them lives in the KMS, never here). Versions are never overwritten.
create table hub.org_keys (
  org_id uuid not null references hub.orgs(id), version int not null, wrapped_dek bytea not null,
  created_at timestamptz not null default now(), primary key (org_id, version)
);

-- ---------------------------------------------------------------------------------------------
-- Identity (global): one account can belong to several organizations
-- ---------------------------------------------------------------------------------------------
create table hub.accounts (
  id uuid primary key default uuidv7(),
  email_idx bytea not null unique,            -- keyed hash: lets us find the account without storing the email in clear
  email_enc bytea not null,                   -- encrypted with the platform data key
  display_name text not null,
  pass_hash text not null,                    -- argon2id
  mfa_enabled boolean not null default false,
  mfa_secret_enc bytea,
  status text not null default 'active' check (status in ('active','inactive','blocked','deleted')),
  failed_attempts int not null default 0,
  locked_until timestamptz,
  password_changed_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);

create table hub.account_orgs (account_id uuid not null references hub.accounts(id), org_id uuid not null references hub.orgs(id),
  primary key (account_id, org_id));

create table hub.security_quiz (
  account_id uuid not null references hub.accounts(id), position smallint not null check (position between 1 and 3),
  question text not null, answer_hash text not null, created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (account_id, position)
);

create table hub.sessions (
  id uuid primary key default uuidv7(), account_id uuid not null references hub.accounts(id),
  org_id uuid references hub.orgs(id), device_id uuid,
  family_id uuid not null,                    -- refresh-token rotation family: reuse of an old token revokes the whole family
  refresh_hash bytea not null unique, expires_at timestamptz not null, revoked_at timestamptz,
  ip inet, user_agent text, created_at timestamptz not null default now(), last_used_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
create index sessions_account_idx on hub.sessions (account_id) where revoked_at is null;

-- Password reset: email -> code -> security questions -> link + second code -> new password.
create table hub.password_resets (
  id uuid primary key default uuidv7(), account_id uuid references hub.accounts(id),
  email_idx bytea not null, stage text not null default 'otp' check (stage in ('otp','quiz','link','done','dead')),
  otp_hash bytea not null, otp_expires timestamptz not null, otp_tries int not null default 0,
  quiz_tries int not null default 0,
  link_hash bytea, link_otp_hash bytea, link_expires timestamptz, link_tries int not null default 0,
  created_at timestamptz not null default now(), ip inet,
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
create index password_resets_email_idx on hub.password_resets (email_idx, created_at desc);

-- ---------------------------------------------------------------------------------------------
-- Organization structure, people, roles (tenant tables: row level security on org_id)
-- ---------------------------------------------------------------------------------------------
create table hub.org_units (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  kind text not null check (kind in ('sector','branch','unit','department')),
  parent_id uuid references hub.org_units(id), name text not null, code text,
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  unique (org_id, kind, name)
);
create index org_units_org_idx on hub.org_units (org_id, kind);

create table hub.people (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  account_id uuid references hub.accounts(id),
  first_name text not null, last_name text not null, title text not null default '',
  status text not null default 'active' check (status in ('active','inactive','blocked')),
  availability text not null default 'Available',
  sector_id uuid references hub.org_units(id), branch_id uuid references hub.org_units(id),
  unit_id uuid references hub.org_units(id), department_id uuid references hub.org_units(id),
  email_idx bytea, email_enc bytea, phone_enc bytea, dob_enc bytea, gender_enc bytea,   -- personal data: encrypted at rest
  color text not null default '#7CC4FF', joined_on date,
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
create index people_org_idx on hub.people (org_id, last_name, first_name);
create index people_dept_idx on hub.people (org_id, department_id);
create unique index people_email_uidx on hub.people (org_id, email_idx) where email_idx is not null;
create unique index people_account_uidx on hub.people (org_id, account_id) where account_id is not null;

create table hub.roles (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  key text not null, name text not null, description text not null default '', scope text not null default 'org',
  locked boolean not null default false, needs_level text,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), unique (org_id, key)
);

create table hub.role_rights (
  org_id uuid not null references hub.orgs(id), role_id uuid not null references hub.roles(id),
  module text not null, right_name text not null, allowed boolean not null,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, role_id, module, right_name)
);

create table hub.person_roles (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id), role_id uuid not null references hub.roles(id),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id)
);

create table hub.person_rights (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  module text not null, right_name text not null, allowed boolean not null,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id, module, right_name)
);

-- Reporting lines: a role reports to a role, optionally a named person; a person may have their own manager.
create table hub.role_reports (
  org_id uuid not null references hub.orgs(id), role_id uuid not null references hub.roles(id),
  reports_to_role uuid references hub.roles(id), reports_to_person uuid references hub.people(id),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, role_id)
);
create table hub.person_reports (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  manager_id uuid not null references hub.people(id),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id)
);

-- ---------------------------------------------------------------------------------------------
-- End-to-end encryption: the key directory and opaque message storage.
-- The server holds public keys, wrapped (already encrypted) group keys and ciphertext. It cannot read any of it.
-- ---------------------------------------------------------------------------------------------
create table hub.devices (
  id uuid primary key default uuidv7(), account_id uuid not null references hub.accounts(id),
  label text not null default '',
  identity_key bytea not null,                -- X25519 public key (key agreement)
  signing_key bytea not null,                 -- Ed25519 public key (signatures)
  signed_prekey bytea not null, signed_prekey_sig bytea not null,
  created_at timestamptz not null default now(), revoked_at timestamptz
);
create index devices_account_idx on hub.devices (account_id) where revoked_at is null;

create table hub.one_time_prekeys (
  device_id uuid not null references hub.devices(id), key_id int not null, public_key bytea not null,
  claimed_at timestamptz, primary key (device_id, key_id)
);

create table hub.conversations (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id),
  kind text not null check (kind in ('dm','forum','wall','note','ticket','call')),
  name text, about text, color text, initials text,        -- metadata the directory needs; content is never here
  privacy text not null default 'invite',
  epoch int not null default 1,                            -- bumps whenever membership changes so removed people lose access to new messages
  created_by uuid references hub.accounts(id), created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
create index conversations_org_idx on hub.conversations (org_id, kind);

create table hub.conversation_members (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  account_id uuid not null references hub.accounts(id), role text not null default 'member',
  epoch int not null, wrapped_key bytea not null,           -- the conversation key, encrypted to this member's device key by a client
  joined_at timestamptz not null default now(), left_at timestamptz,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, conversation_id, account_id, epoch)
);
create index conversation_members_account_idx on hub.conversation_members (account_id, org_id) where left_at is null;

create table hub.messages (
  org_id uuid not null, id uuid not null default uuidv7(), conversation_id uuid not null,
  sender_account uuid not null, sender_device uuid, epoch int not null,
  ciphertext bytea not null, content_type text not null default 'text',
  reply_to uuid, created_at timestamptz not null default now(), deleted_at timestamptz,
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id)
) partition by hash (org_id);
do $$ begin
  for i in 0..31 loop
    execute format('create table hub.messages_p%s partition of hub.messages for values with (modulus 32, remainder %s)', i, i);
  end loop;
end $$;
create index messages_conv_idx on hub.messages (org_id, conversation_id, id desc);

-- ---------------------------------------------------------------------------------------------
-- Tamper-evident ledger: one hash chain per organization. Each entry commits to the previous one, so removing or
-- changing any past entry breaks every hash after it. Checkpoints are signed with a key held outside the database.
-- ---------------------------------------------------------------------------------------------
create table hub.ledger_heads (org_id uuid primary key, seq bigint not null, hash bytea not null);
create table hub.ledger (
  org_id uuid not null, seq bigint not null, ts timestamptz not null, actor uuid, action text not null, subject text not null default '',
  payload_hash bytea not null, prev_hash bytea not null, hash bytea not null, sig bytea not null,
  primary key (org_id, seq)
);
create table hub.ledger_checkpoints (
  org_id uuid not null, upto_seq bigint not null, root bytea not null, sig bytea not null, created_at timestamptz not null default now(),
  primary key (org_id, upto_seq)
);

-- Queryable activity log shown on the Security & logs page. Each row points at its ledger entry.
create table hub.audit_logs (
  org_id uuid not null, id uuid not null default uuidv7(), ts timestamptz not null default now(),
  actor_name text not null, action text not null, module text not null, ip text not null default '', location text not null default '',
  result text not null default 'Success', rule text not null default '', ledger_seq bigint,
  primary key (org_id, id)
);
create index audit_logs_ts_idx on hub.audit_logs (org_id, ts desc);

-- ---------------------------------------------------------------------------------------------
-- Reference data seeded for every client (not tenant data)
-- ---------------------------------------------------------------------------------------------
create table hub.tones (
  key text primary key, kind text not null, description text not null, params jsonb not null,
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);

-- Wire up the guards
select hub.protect('hub.orgs', false);
select hub.protect('hub.org_keys', false, 'readonly');
select hub.protect('hub.accounts', false);
select hub.protect('hub.account_orgs', false, 'guard');
select hub.protect('hub.security_quiz', false);
select hub.protect('hub.sessions', false);
select hub.protect('hub.password_resets', false);
select hub.protect('hub.org_units');
select hub.protect('hub.people');
select hub.protect('hub.roles');
select hub.protect('hub.role_rights');
select hub.protect('hub.person_roles');
select hub.protect('hub.person_rights');
select hub.protect('hub.role_reports');
select hub.protect('hub.person_reports');
select hub.protect('hub.devices', false, 'guard');
select hub.protect('hub.one_time_prekeys', false, 'guard');
select hub.protect('hub.conversations');
select hub.protect('hub.conversation_members');
select hub.protect('hub.messages');
select hub.protect('hub.ledger_heads', false, 'guard');
select hub.protect('hub.ledger', false, 'readonly');
select hub.protect('hub.ledger_checkpoints', false, 'readonly');
select hub.protect('hub.audit_logs', false, 'readonly');
select hub.protect('hub.tones', false);
