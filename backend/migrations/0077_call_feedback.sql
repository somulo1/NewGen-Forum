-- Feedback on a call, meeting or standup: each person who was in it (a call nobody joined has none) rates it 1 to 5 stars once,
-- with an optional remark sealed with the organization's key. Final once given.
create table hub.call_feedback (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), call_id uuid not null,
  person_id uuid not null references hub.people(id), stars smallint not null check (stars between 1 and 5), remark_enc bytea,
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id), unique (org_id, call_id, person_id)
);
create index call_feedback_call_idx on hub.call_feedback (org_id, call_id);
select hub.protect('hub.call_feedback');
