-- Push notifications for the Hub phone app. A phone registers the address its platform gives it (an FCM token on Android, an
-- APNs device token on iOS) for one sign-in: pushes go to it only while that sign-in is alive and in that organization, so a
-- phone that signs out, or whose sign-in is replaced elsewhere, stops receiving them at once. The address is kept encrypted and
-- looked up by its hash.
create table hub.push_devices (
  token_hash bytea primary key check (length(token_hash) = 32),
  account_id uuid not null references hub.accounts(id),
  family_id uuid not null,
  platform text not null check (platform in ('android', 'ios')),
  token_enc bytea not null,
  created_at timestamptz not null default now(),
  seen_at timestamptz not null default now()
);
create index push_devices_account_idx on hub.push_devices (account_id);
select hub.protect('hub.push_devices', false, 'guard');

-- pushes check that the sign-in is still alive
create index if not exists sessions_family_idx on hub.sessions (family_id) where revoked_at is null;
