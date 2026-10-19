-- Where an award should be announced (the community wall and/or forums) and whether that has been done. The nominator's
-- browser writes the (end-to-end encrypted) posts, because only members can; claiming a target first means it is posted once.
create table hub.award_posts (
  org_id uuid not null references hub.orgs(id), award_id uuid not null, target text not null,
  claimed boolean not null default false,
  ver int not null default 1, mac bytea not null default hub.pending_mac(), primary key (org_id, award_id, target)
);
select hub.protect('hub.award_posts');
