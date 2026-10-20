-- Email details are no longer kept: they travel in the request that causes the email and are never stored.
drop table if exists hub.email_details;
