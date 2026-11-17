-- The phone each account is signed in on (the key the Hub app keeps in that phone's secure storage). Signing in on another
-- phone needs a code emailed to the account; once confirmed, that phone replaces this one and the old phone is signed out.
create table hub.account_phones (
  account_id uuid primary key references hub.accounts(id),
  public_key bytea not null check (length(public_key) = 32),
  platform text not null default '',
  bound_at timestamptz not null default now()
);
select hub.protect('hub.account_phones', false, 'guard');
