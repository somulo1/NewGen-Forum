-- A temporary password set by an administrator must be replaced at the next sign-in.
alter table hub.accounts add column must_change boolean not null default false;

-- Free-text parts of a profile. Everything here is personal, so it is encrypted with the organization's data key.
create table hub.person_profiles (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  bio_enc bytea, skills_enc bytea, photo_enc bytea, photo_type text, cover text not null default '#1F3A2B',
  updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id)
);

-- Invitations to join an organization: the token is stored only as a hash and works once.
create table hub.person_invites (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  token_hash bytea not null, expires_at timestamptz not null, used_at timestamptz,
  created_by uuid references hub.people(id), created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id, token_hash)
);
create index person_invites_token_idx on hub.person_invites (token_hash);

select hub.protect('hub.person_profiles');
select hub.protect('hub.person_invites');
