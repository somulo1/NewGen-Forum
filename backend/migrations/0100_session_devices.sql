-- Sign-ins made from the Hub phone app are bound to that phone: the app keeps a private key in the phone's secure storage and
-- sends the public half when signing in. Every refresh of that sign-in must be signed with the same key (see auth/device.go).
-- One row per sign-in (refresh-token family); a binding never changes, so rows can be added but not edited or removed.
create table hub.session_devices (
  family_id uuid primary key, account_id uuid not null references hub.accounts(id),
  public_key bytea not null check (length(public_key) = 32), platform text not null default '',
  created_at timestamptz not null default now()
);
create index session_devices_account_idx on hub.session_devices (account_id);
select hub.protect('hub.session_devices', false, 'readonly');
