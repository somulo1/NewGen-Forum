-- The days each person used the system (signed in or was active), one row per person per day however often they come back.
create table hub.person_days (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id), day date not null,
  primary key (org_id, person_id, day)
);
create index person_days_day_idx on hub.person_days (org_id, day);
select hub.protect('hub.person_days', true, 'guard');
