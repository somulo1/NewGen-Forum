-- Files. The bytes are ciphertext in the blob store; the name, type and file key are inside `envelope`, which only the
-- folder's members can open.
create table hub.files (
  org_id uuid not null references hub.orgs(id), id uuid not null, folder_id uuid not null references hub.conversations(id),
  uploader uuid not null references hub.people(id), author_account uuid not null, author_device uuid not null,
  epoch int not null, envelope bytea not null, size bigint, sha256 bytea, blob_key text,
  status text not null default 'uploading' check (status in ('uploading','ready','deleted')),
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, id)
);
create index files_folder_idx on hub.files (org_id, folder_id, created_at desc) where status = 'ready';

-- Links for people outside Hub. The token is stored only as a hash; the key to open the file lives in the link's #fragment,
-- which a browser never sends to any server. `key_envelope` is the file's key and metadata sealed under that link key.
create table hub.shares (
  org_id uuid not null references hub.orgs(id), id uuid not null default uuidv7(), file_id uuid not null,
  token_hash bytea not null unique, created_by uuid not null references hub.people(id),
  status text not null default 'Active' check (status in ('Pending approval','Active','Revoked')),
  expires_at timestamptz not null, max_downloads int not null default 1, downloads int not null default 0,
  pass_hash text, key_envelope bytea not null, created_at timestamptz not null default now(),
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, id)
);
create index shares_file_idx on hub.shares (org_id, file_id);
select hub.protect('hub.files');
select hub.protect('hub.shares');
