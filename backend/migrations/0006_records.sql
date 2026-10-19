-- More kinds of encrypted container, audiences for the wall, and a flag for "someone left, rotate the key".
alter table hub.conversations drop constraint conversations_kind_check;
alter table hub.conversations add constraint conversations_kind_check check (kind in ('dm','forum','wall','note','ticket','call','project','event','folder'));
alter table hub.conversations add column audience text check (audience in ('org','sector','branch','unit','department')), add column audience_ref uuid,
  add column rotate_needed boolean not null default false;
create index conversations_audience_idx on hub.conversations (org_id, audience, audience_ref) where audience is not null;

-- Encrypted records: a note, a task, a milestone, a calendar entry, a ticket comment ... The content is a client-made
-- envelope under the container's conversation key. `meta` holds the few plain fields the server must see to do its job
-- (a task's board column, a due date); each collection says which fields it may carry.
create table hub.records (
  org_id uuid not null references hub.orgs(id), id uuid not null,
  collection text not null, container_id uuid not null,
  parent_id uuid, author_account uuid not null, author_person uuid, author_device uuid,
  epoch int not null, ciphertext bytea not null, meta jsonb not null default '{}',
  created_at timestamptz not null default now(), updated_at timestamptz not null default now(), deleted_at timestamptz,
  ver int not null default 1, mac bytea not null default hub.pending_mac(),
  primary key (org_id, id)
) partition by hash (org_id);
do $$ begin
  for i in 0..15 loop
    execute format('create table hub.records_p%s partition of hub.records for values with (modulus 16, remainder %s)', i, i);
  end loop;
end $$;
create index records_container_idx on hub.records (org_id, container_id, collection, created_at desc);
create index records_parent_idx on hub.records (org_id, parent_id) where parent_id is not null;
create index records_meta_idx on hub.records using gin (meta jsonb_path_ops);
select hub.protect('hub.records');
