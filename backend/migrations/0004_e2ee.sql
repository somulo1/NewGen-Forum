-- End-to-end encryption, round two: keys are wrapped per DEVICE (so a person's phone and laptop each hold their own copy),
-- and membership and keys are separate things.

drop table hub.one_time_prekeys;
alter table hub.devices drop column signed_prekey, drop column signed_prekey_sig;
alter table hub.devices add column binding_sig bytea not null default '\x', add column last_seen_at timestamptz;
alter table hub.devices add constraint devices_keys_len check (length(identity_key) = 32 and length(signing_key) = 32);
create unique index devices_identity_uidx on hub.devices (identity_key);

-- A password-protected copy of a person's private keys, encrypted ON THE DEVICE with a key made from a recovery passphrase
-- the server never sees. It lets someone sign in on a new phone without losing their history.
create table hub.key_backups (
  account_id uuid primary key references hub.accounts(id), blob bytea not null, kdf jsonb not null,
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac()
);
select hub.protect('hub.key_backups', false);

-- Membership: who is in the conversation, and from which key epoch they may read (history before that stays closed to them).
alter table hub.conversation_members drop constraint conversation_members_pkey;
alter table hub.conversation_members drop column epoch, drop column wrapped_key;
alter table hub.conversation_members add column first_epoch int not null default 1, add column muted boolean not null default false;
alter table hub.conversation_members add primary key (org_id, conversation_id, account_id);
alter table hub.conversations add column pair_key text, add column pinned_message uuid;
create unique index conversations_pair_uidx on hub.conversations (org_id, pair_key) where pair_key is not null;
alter table hub.conversations drop constraint conversations_kind_check;
alter table hub.conversations add constraint conversations_kind_check check (kind in ('dm','forum','wall','note','ticket','call','project'));

-- The wrapped conversation keys: one row per (conversation, epoch, device). They are ciphertext: only that device can open them.
create table hub.conversation_keys (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  epoch int not null, device_id uuid not null references hub.devices(id),
  wrapped_key bytea not null, wrapped_by uuid references hub.devices(id), created_at timestamptz not null default now(),
  primary key (org_id, conversation_id, epoch, device_id)
);
create index conversation_keys_device_idx on hub.conversation_keys (device_id);
select hub.protect('hub.conversation_keys', true, 'readonly');

-- Where each member has read up to. Small, hot, and only meaningful to its owner, so it is guarded rather than MAC-sealed.
create table hub.read_cursors (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  account_id uuid not null references hub.accounts(id), last_read uuid, updated_at timestamptz not null default now(),
  primary key (org_id, conversation_id, account_id)
);
select hub.protect('hub.read_cursors', true, 'guard');

alter table hub.messages add column sender_person uuid;
