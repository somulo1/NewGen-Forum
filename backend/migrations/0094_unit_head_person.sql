-- The person who heads a place. Each level has a head role (Organization structure > levels); for each place, one person is
-- chosen to lead it: someone holding that role, or one of the place's members. They are placed in it when chosen. While no one
-- is chosen, everyone placed there who holds the head role heads it, as before.
create table hub.unit_head_people (
  org_id uuid not null references hub.orgs(id), unit_id uuid not null references hub.org_units(id) on delete cascade,
  person_id uuid not null references hub.people(id), updated_at timestamptz not null default now(),
  primary key (org_id, unit_id)
);
select hub.protect('hub.unit_head_people', true, 'guard');

-- the chosen head of a place, or null when none was chosen (then the head role decides)
create function hub.chosen_head(u uuid) returns uuid language sql stable as $$
  select person_id from hub.unit_head_people where unit_id = u
$$;
