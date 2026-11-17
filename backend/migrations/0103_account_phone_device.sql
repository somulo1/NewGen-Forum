-- The phone as its system names it for the Hub app (a keyed hash of a hash; the phone's own identifier never reaches the server).
-- Reinstalling or updating the app on the same phone makes a new key but keeps this, so it is not taken for a new phone.
alter table hub.account_phones add column device_hash bytea;
