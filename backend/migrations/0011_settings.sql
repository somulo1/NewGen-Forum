-- Organization settings by section (profile, subscription, billing, security). Each section is one JSON document encrypted
-- with the organization's data key, bound to its row.
create table hub.org_settings (
  org_id uuid not null references hub.orgs(id), section text not null check (section in ('profile','subscription','billing','security')),
  data_enc bytea not null, updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, section)
);
create table hub.invoices (
  org_id uuid not null references hub.orgs(id), no text not null, issued_on date not null, amount_minor bigint not null check (amount_minor >= 0),
  currency text not null default 'KES', status text not null check (status in ('Paid','Due','Void')),
  created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, no)
);
select hub.protect('hub.org_settings');
select hub.protect('hub.invoices');
