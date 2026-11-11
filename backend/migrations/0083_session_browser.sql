-- Each browser keeps a random token and sends it when signing in; only its hash is kept here. A sign-in from a browser the account
-- has used before (whatever the network or place) sends no "new sign-in" email.
alter table hub.sessions add column browser_hash bytea;
create index sessions_browser_idx on hub.sessions (account_id, browser_hash) where browser_hash is not null;
