-- When an award was approved. The Performance awards board on Home shows the awards approved in the current weekly cycle
-- (it starts afresh every Saturday at 00:00 and waits for the next award), so the moment of approval is recorded.
-- A side table (not a new column) so existing sealed award rows stay valid.
create table hub.award_approvals (
  org_id uuid not null references hub.orgs(id), award_id uuid not null, approved_at timestamptz not null default now(),
  primary key (org_id, award_id)
);
select hub.protect('hub.award_approvals', true, 'guard');
