-- Scale: counts kept as running totals instead of being recounted on every page.

-- How many current members each conversation has (a company-wide wall can have millions). Kept up to date by a trigger on
-- conversation_members, so the chat list and communication reach read one row instead of counting every member.
create table hub.conversation_counts (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id) on delete cascade,
  members int not null default 0, primary key (org_id, conversation_id)
);
insert into hub.conversation_counts (org_id, conversation_id, members)
  select org_id, conversation_id, count(*) from hub.conversation_members where left_at is null group by org_id, conversation_id;
select hub.protect('hub.conversation_counts', true, 'guard');

create function hub.count_members() returns trigger language plpgsql as $$
declare d int := 0; o uuid; c uuid;
begin
  if tg_op = 'INSERT' then
    if new.left_at is null then d := 1; end if; o := new.org_id; c := new.conversation_id;
  elsif tg_op = 'DELETE' then
    if old.left_at is null then d := -1; end if; o := old.org_id; c := old.conversation_id;
  else
    if old.left_at is null and new.left_at is not null then d := -1;
    elsif old.left_at is not null and new.left_at is null then d := 1; end if;
    o := new.org_id; c := new.conversation_id;
  end if;
  if d > 0 then
    insert into hub.conversation_counts as cc (org_id, conversation_id, members) values (o, c, d)
      on conflict (org_id, conversation_id) do update set members = cc.members + d;
  elsif d < 0 then
    -- only lower an existing count: erasing an organization never brings a count row back
    update hub.conversation_counts set members = greatest(members + d, 0) where org_id = o and conversation_id = c;
  end if;
  return null;
end $$;
create trigger conversation_members_count after insert or update of left_at or delete on hub.conversation_members
  for each row execute function hub.count_members();

-- The directory is listed in name order and searched by name: an index in that order, and a trigram index for "contains".
create index if not exists people_name_order_idx on hub.people (org_id, first_name, last_name, id);

-- Emails to a whole audience go through an outbox: one row per recipient, sent by the worker in small batches and retried, so a
-- post to millions of people reaches all of them (nothing is capped or held in memory) and a restart loses nothing.
create table hub.mail_jobs (
  id uuid primary key default uuidv7(), org_id uuid not null references hub.orgs(id), subject text not null, card jsonb not null,
  created_at timestamptz not null default now()
);
create table hub.mail_outbox (
  job_id uuid not null references hub.mail_jobs(id) on delete cascade, org_id uuid not null references hub.orgs(id),
  person_id uuid not null, attempts int not null default 0, next_at timestamptz not null default now(),
  primary key (job_id, person_id)
);
create index mail_outbox_due_idx on hub.mail_outbox (next_at);
select hub.protect('hub.mail_jobs', false, 'guard');
select hub.protect('hub.mail_outbox', false, 'guard');

-- "Name contains" search uses trigram indexes (pg_trgm, available on Neon and standard Postgres). Where the extension cannot be
-- turned on, the migration still succeeds: search works, only without these two indexes.
do $$
begin
  create extension if not exists pg_trgm;
  create index if not exists people_name_trgm_idx on hub.people using gin (lower(first_name || ' ' || last_name) gin_trgm_ops);
  create index if not exists people_title_trgm_idx on hub.people using gin (lower(title) gin_trgm_ops);
exception when insufficient_privilege or undefined_file or feature_not_supported then
  raise notice 'pg_trgm is not available here; people search runs without trigram indexes';
end $$;
