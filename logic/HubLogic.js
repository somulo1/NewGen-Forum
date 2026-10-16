import { DCLogic } from '~/utils/dcLogic.js';
import { coreState } from '~/logic/state/core.js';
import { awardsState } from '~/logic/state/awards.js';
import { calendarState } from '~/logic/state/calendar.js';
import { commsState } from '~/logic/state/comms.js';
import { adminState } from '~/logic/state/admin.js';
import { profileState } from '~/logic/state/profile.js';
import { filesState } from '~/logic/state/files.js';
import { projectsState } from '~/logic/state/projects.js';
import { callsState } from '~/logic/state/calls.js';
import { settingsState } from '~/logic/state/settings.js';
import { ticketsState } from '~/logic/state/tickets.js';
import { wallState } from '~/logic/state/wall.js';
import { forumsState } from '~/logic/state/forums.js';
import { onboardingState } from '~/logic/state/onboarding.js';
import { seedProjects } from '~/logic/seed/projects.js';
import { seedOrg } from '~/logic/seed/org.js';
import { seedTickets } from '~/logic/seed/tickets.js';
import { seedDirectory } from '~/logic/seed/directory.js';
import { seedLogs } from '~/logic/seed/logs.js';
import { messagingActions } from '~/logic/actions/messaging.js';
import { projectsActions } from '~/logic/actions/projects.js';
import { callsActions } from '~/logic/actions/calls.js';
import { uiActions } from '~/logic/actions/ui.js';
import { forumActions } from '~/logic/actions/forums.js';
import { onboardingActions } from '~/logic/actions/onboarding.js';
import { ticketCatActions } from '~/logic/actions/ticketCats.js';
import { buildAccess } from '~/logic/views/access.js';
import { buildForums } from '~/logic/views/forums.js';
import { buildNav } from '~/logic/views/nav.js';
import { buildWall } from '~/logic/views/wall.js';
import { buildTickets } from '~/logic/views/tickets.js';
import { buildProjects } from '~/logic/views/projects.js';
import { buildPeople } from '~/logic/views/people.js';
import { buildSecurity } from '~/logic/views/security.js';
import { buildRoles } from '~/logic/views/roles.js';
import { buildTableHelpers } from '~/logic/views/tableHelpers.js';
import { buildCalendar } from '~/logic/views/calendar.js';
import { buildOnboarding } from '~/logic/views/onboarding.js';
import { buildForumExtras } from '~/logic/views/forumExtras.js';
import { buildCatMgr } from '~/logic/views/catMgr.js';
import { buildProjectTables } from '~/logic/views/projectTables.js';
import { buildProjectStandup } from '~/logic/views/projectStandup.js';
import { buildOrg } from '~/logic/views/org.js';
import { buildComms } from '~/logic/views/comms.js';
import { buildAdminTables } from '~/logic/views/adminTables.js';
import { buildAwards } from '~/logic/views/awards.js';
import { buildHomeWidgets } from '~/logic/views/homeWidgets.js';
import { buildProfile } from '~/logic/views/profile.js';
import { buildWidgetSettings } from '~/logic/views/widgetSettings.js';
import { buildCallState } from '~/logic/views/callState.js';
import { valsAccess } from '~/logic/vals/access.js';
import { buildSettings } from '~/logic/views/settings.js';
import { valsWall } from '~/logic/vals/wall.js';
import { valsPages } from '~/logic/vals/pages.js';
import { valsProjects } from '~/logic/vals/projects.js';
import { valsTickets } from '~/logic/vals/tickets.js';
import { valsShell } from '~/logic/vals/shell.js';
import { valsHome } from '~/logic/vals/home.js';
import { valsPanels } from '~/logic/vals/panels.js';

// Hub logic. The pieces live in logic/:
//   state/    initial state, grouped by area
//   seed/     sample data derived from the state when the app starts
//   actions/  methods mixed into HubLogic
//   views/    renderVals() steps, run in order; each returns the names later steps need
//   vals/     the values and handlers the templates bind to
//   projects.js  milestone, objective, task and sprint rules
export class HubLogic extends DCLogic {
  constructor(props) {
    super(props);
    this.state = Object.assign({}, coreState(), awardsState(), calendarState(), commsState(), adminState(), profileState(), filesState(), projectsState(), callsState(), ticketsState(), wallState(), settingsState(), forumsState(), onboardingState());
    seedProjects.call(this);
    seedOrg.call(this);
    seedTickets.call(this);
    seedDirectory.call(this);
    seedLogs.call(this);
  }
  setState(update, done) {
    const raw = typeof update === 'function' ? update(this.state, this.props) : update;
    const patch = raw ? Object.assign({}, raw) : raw;
    if (patch) {
      const person = patch.personOpen !== undefined && patch.personOpen !== null ? this.dir[patch.personOpen] : null;
      if (person && this._canFullProfile) {
        patch.personOpen = null; patch.profileOf = person.name; patch.section = 'profile';
        patch.notifOpen = false; patch.statusMenu = false; patch.profileEdit = false;
        if (this.state.mode === 'full' || this.state.mode === 'page') patch.mode = 'docked';
      } else if (patch.section !== undefined && patch.profileOf === undefined) {
        patch.profileOf = null;
      }
    }
    super.setState(patch, done);
  }
  renderVals() {
    const ctx = {};
    for (const step of VIEW_STEPS) Object.assign(ctx, step.call(this, ctx));
    const out = {};
    for (const part of VAL_PARTS) Object.assign(out, part.call(this, ctx));
    return out;
  }
}

Object.assign(HubLogic.prototype, messagingActions, projectsActions, callsActions, uiActions, forumActions, onboardingActions, ticketCatActions);

const VIEW_STEPS = [buildAccess, buildForums, buildNav, buildWall, buildTickets, buildForumExtras, buildProjects, buildPeople, buildOnboarding, buildSecurity, buildRoles, buildTableHelpers, buildProjectTables, buildCalendar, buildProjectStandup, buildOrg, buildCatMgr, buildComms, buildAdminTables, buildSettings, buildAwards, buildHomeWidgets, buildProfile, buildWidgetSettings, buildCallState];
const VAL_PARTS = [valsAccess, valsPages, valsProjects, valsTickets, valsShell, valsHome, valsWall, valsPanels];
