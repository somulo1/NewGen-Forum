-- Who may post in a forum (everyone, certain roles, or an allow-list), with people who are blocked from posting.
create table hub.forum_settings (
  org_id uuid not null references hub.orgs(id), conversation_id uuid not null references hub.conversations(id),
  post_mode text not null default 'everyone' check (post_mode in ('everyone','roles','list')),
  roles text[] not null default '{}', blocked uuid[] not null default '{}', allowed uuid[] not null default '{}',
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, conversation_id)
);
select hub.protect('hub.forum_settings');

-- A person's own settings for the app (which widgets are on their home screen, theme ...). Encrypted at rest.
create table hub.person_prefs (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  data_enc bytea not null, updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id)
);
select hub.protect('hub.person_prefs');
