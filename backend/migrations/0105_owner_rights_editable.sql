-- The owner role's rights can be changed like any role's. It holds a right until it is switched off for it (a right added to the
-- catalog later is on for it at once); every other role holds a right only once it is switched on.
create or replace function hub.role_right(p_role uuid, p_module text, p_right text) returns boolean language sql stable as $$
  select coalesce((select rr.allowed from hub.role_rights rr where rr.role_id = p_role and rr.module = p_module and rr.right_name = p_right),
                  (select r.locked from hub.roles r where r.id = p_role), false)
$$;
