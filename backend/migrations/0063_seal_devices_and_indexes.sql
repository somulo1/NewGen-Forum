-- 1. Two security-critical tables had no per-row seal: the devices (their public keys decide who can read a conversation) and
--    who belongs to which organization. They are sealed like the rest, so a change made behind the application's back shows up in
--    the integrity scan. Existing rows start pending; `bootstrap seal` seals them once.
alter table hub.devices add column if not exists ver int not null default 1, add column if not exists mac bytea not null default hub.pending_mac();
alter table hub.account_orgs add column if not exists ver int not null default 1, add column if not exists mac bytea not null default hub.pending_mac();
create trigger hub_devices_mac before insert or update on hub.devices for each row execute function hub.require_mac();
create trigger hub_account_orgs_mac before insert or update on hub.account_orgs for each row execute function hub.require_mac();

-- 2. Indexes for the foreign keys and lookups that had none (a delete or a join on them scanned the whole table).
create index if not exists ticket_assignees_person_idx on hub.ticket_assignees (org_id, person_id);
create index if not exists tickets_raised_by_idx on hub.tickets (org_id, raised_by);
create index if not exists comms_conversation_idx on hub.comms (org_id, conversation_id);
create index if not exists comms_by_person_idx on hub.comms (org_id, by_person);
create index if not exists calls_conversation_idx on hub.calls (org_id, conversation_id);
create index if not exists calls_host_idx on hub.calls (org_id, host);
create index if not exists read_cursors_account_idx on hub.read_cursors (org_id, account_id);
create index if not exists approval_requests_requester_idx on hub.approval_requests (org_id, requester);
create index if not exists approval_steps_approver_idx on hub.approval_steps (org_id, approver);
create index if not exists approval_designees_person_idx on hub.approval_designees (org_id, person_id);
create index if not exists alert_notes_person_idx on hub.alert_notes (org_id, person);
create index if not exists awards_nominated_by_idx on hub.awards (org_id, nominated_by);
create index if not exists comm_reads_person_idx on hub.comm_reads (org_id, person_id);
create index if not exists people_sector_idx on hub.people (org_id, sector_id);
create index if not exists people_branch_idx on hub.people (org_id, branch_id);
create index if not exists people_unit_idx on hub.people (org_id, unit_id);
create index if not exists person_roles_role_idx on hub.person_roles (org_id, role_id);
create index if not exists person_reports_manager_idx on hub.person_reports (org_id, manager_id);
create index if not exists role_reports_to_role_idx on hub.role_reports (org_id, reports_to_role);
create index if not exists role_reports_to_person_idx on hub.role_reports (org_id, reports_to_person);
create index if not exists org_units_parent_idx on hub.org_units (org_id, parent_id);
create index if not exists projects_owner_idx on hub.projects (org_id, owner);
create index if not exists shares_created_by_idx on hub.shares (org_id, created_by);
create index if not exists person_invites_created_by_idx on hub.person_invites (org_id, created_by);
create index if not exists default_role_role_idx on hub.default_role (org_id, role_id);
create index if not exists sessions_org_idx on hub.sessions (org_id) where org_id is not null;
create index if not exists password_resets_account_idx on hub.password_resets (account_id);
create index if not exists login_blocks_account_idx on hub.login_blocks (account_id) where account_id is not null;
create index if not exists conversations_created_by_idx on hub.conversations (created_by);
