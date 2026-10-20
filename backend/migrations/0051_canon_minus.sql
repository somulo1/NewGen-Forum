-- The same canonical form as hub.canon, without some columns. The integrity check uses it to tell a record that was edited
-- outside the system from one that only looks different because a database update added a column after it was sealed.
create function hub.canon_minus(r anyelement, drop_keys text[]) returns text language sql stable set timezone = 'UTC' as
$$ select (to_jsonb(r) - 'mac' - 'ver' - 'updated_at' - drop_keys)::text $$;
