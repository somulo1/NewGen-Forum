-- A record of every time sign-in was blocked, kept for the Security page. Blocks lift by themselves; the record stays.
-- kind: 'account' (too many wrong passwords for one account) or 'address' (too many attempts from one network address).
-- An account is null when the address tried emails that belong to nobody.
create table hub.login_blocks (
  id uuid primary key default uuidv7(), at timestamptz not null default now(),
  kind text not null check (kind in ('account','address')),
  account_id uuid references hub.accounts(id), email_hint text not null default '', ip text not null default '',
  reason text not null, until timestamptz not null
);
create index login_blocks_at_idx on hub.login_blocks (at desc);
select hub.protect('hub.login_blocks', false, 'readonly');
