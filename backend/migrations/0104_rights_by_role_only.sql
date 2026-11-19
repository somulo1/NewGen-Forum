-- A role holds a right only through its own setting: user groups no longer add to it. The API folds every user group's rights
-- into its roles when it starts (widgetaccess.FoldUserGroups), so nobody loses what they could do; the group tables are then empty.
create or replace function hub.role_right(p_role uuid, p_module text, p_right text) returns boolean language sql stable as $$
  select coalesce((select rr.allowed from hub.role_rights rr where rr.role_id = p_role and rr.module = p_module and rr.right_name = p_right), false)
$$;
