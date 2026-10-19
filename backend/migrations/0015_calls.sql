-- Call history. Media never touches this database (it flows between devices, encrypted); only who called whom and for
-- how long is kept.
create table hub.calls (
  org_id uuid not null references hub.orgs(id), id uuid not null, kind text not null check (kind in ('voice','video')),
  context text not null check (context in ('dm','forum','meeting','project','standup')), conversation_id uuid references hub.conversations(id),
  host uuid not null references hub.people(id), title_enc bytea,
  status text not null default 'ringing' check (status in ('ringing','active','ended','missed','declined')),
  started_at timestamptz not null default now(), answered_at timestamptz, ended_at timestamptz,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, id)
);
create index calls_recent_idx on hub.calls (org_id, started_at desc);
create table hub.call_participants (
  org_id uuid not null references hub.orgs(id), call_id uuid not null, person_id uuid not null references hub.people(id),
  invited boolean not null default false, joined_at timestamptz, left_at timestamptz,
  primary key (org_id, call_id, person_id)
);
create index call_participants_person_idx on hub.call_participants (org_id, person_id, joined_at desc);
select hub.protect('hub.calls');
select hub.protect('hub.call_participants', true, 'guard');
