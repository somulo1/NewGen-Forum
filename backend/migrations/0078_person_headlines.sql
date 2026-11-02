-- A person's professional headline, like "Software developer | DevOps", shown next to their name on the community wall and in the directory.
-- They write it themselves on their profile page. It is its own table (not a column on people), so the seals of existing rows stay valid.
create table hub.person_headlines (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  headline text not null check (length(headline) between 1 and 120),
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id)
);
select hub.protect('hub.person_headlines');
