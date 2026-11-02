-- Which structure levels an organization uses (sector, branch, unit, department), what it calls them and which sits under which are
-- saved as the settings section 'structure'.
alter table hub.org_settings drop constraint org_settings_section_check;
alter table hub.org_settings add constraint org_settings_section_check check (section in ('profile','subscription','billing','security','widgets','structure'));
