-- Someone who closes the rating card without rating is not asked about that call again.
create table hub.call_feedback_skips (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), call_id uuid not null,
  person_id uuid not null references hub.people(id),
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id), unique (org_id, call_id, person_id)
);
select hub.protect('hub.call_feedback_skips');
