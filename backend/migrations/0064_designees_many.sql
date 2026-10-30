-- A level of an approval flow can have several chosen people (any one of them may decide it). The key now includes the person.
alter table hub.approval_designees drop constraint approval_designees_pkey;
alter table hub.approval_designees add primary key (org_id, flow_key, level, person_id);
