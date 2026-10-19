-- Anyone with a share link has no account, so nothing tells us which organization it belongs to. This lets one narrow lookup
-- find a link by the hash of its token (rows are visible only to a caller that presents that hash) and returns just what
-- serving the link needs.
create function hub.share_hash() returns bytea language sql stable as
$$ select decode(nullif(current_setting('hub.share_hash', true), ''), 'hex') $$;
create policy by_token on hub.shares for select using (token_hash = hub.share_hash());

create function hub.share_lookup(h bytea) returns table (org_id uuid, share_id uuid, file_id uuid, key_envelope bytea, size bigint, pass_hash text, blob_key text)
language plpgsql security definer set search_path = hub, public as $$
declare o uuid;
begin
  perform set_config('hub.share_hash', encode(h, 'hex'), true);
  select s.org_id into o from hub.shares s where s.token_hash = h and s.status = 'Active' and s.expires_at > now() and s.downloads < s.max_downloads;
  if o is null then return; end if;
  perform set_config('hub.org_id', o::text, true);
  return query select s.org_id, s.id, s.file_id, s.key_envelope, f.size, s.pass_hash, f.blob_key
    from hub.shares s join hub.files f on f.id = s.file_id and f.org_id = s.org_id
    where s.token_hash = h and f.status = 'ready';
  perform set_config('hub.org_id', '', true);
end $$;
revoke all on function hub.share_lookup(bytea) from public;
