-- The organization's home-widget settings (names, who sees and manages each widget) are saved as section 'widgets', which the
-- original check did not allow, so every save failed with a server error.
alter table hub.org_settings drop constraint org_settings_section_check;
alter table hub.org_settings add constraint org_settings_section_check check (section in ('profile','subscription','billing','security','widgets'));
