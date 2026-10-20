-- How much one person may store (MB, 0 = no limit), when it differs from the organization's default in security settings.
create table hub.storage_quotas (
  org_id uuid not null references hub.orgs(id), person_id uuid not null references hub.people(id),
  quota_mb bigint not null check (quota_mb >= 0), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, person_id)
);
select hub.protect('hub.storage_quotas');

-- what each person has stored, summed on every upload
create index if not exists files_uploader_idx on hub.files (org_id, uploader) where status = 'ready';
