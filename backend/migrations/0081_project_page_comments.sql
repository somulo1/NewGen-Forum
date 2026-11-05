-- Comments on a project's documents.
alter table hub.project_items drop constraint project_items_collection_check;
alter table hub.project_items add constraint project_items_collection_check check (collection in ('task','milestone','sprint','risk','activity','page','page_comment'));
