import { DCLogic } from '~/utils/dcLogic.js';
import { stopRing, stopRingback } from '~/logic/tones.js';
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
import { notesState } from '~/logic/state/notes.js';
import { buildNotes } from '~/logic/views/notes.js';
import { ticketsState } from '~/logic/state/tickets.js';
import { wallState } from '~/logic/state/wall.js';
import { forumsState } from '~/logic/state/forums.js';
import { onboardingState } from '~/logic/state/onboarding.js';
import { messagingActions } from '~/logic/actions/messaging.js';
import { projectsActions } from '~/logic/actions/projects.js';
import { callsActions } from '~/logic/actions/calls.js';
import { uiActions } from '~/logic/actions/ui.js';
import { forumActions } from '~/logic/actions/forums.js';
import { onboardingActions } from '~/logic/actions/onboarding.js';
import { peopleAdminActions } from '~/logic/actions/peopleAdmin.js';
import { ticketCatActions } from '~/logic/actions/ticketCats.js';
import { noteActions } from '~/logic/actions/notes.js';
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
import { buildReporting, buildReportingTab } from '~/logic/views/reporting.js';
import { valsWall } from '~/logic/vals/wall.js';
import { valsPages } from '~/logic/vals/pages.js';
import { valsProjects } from '~/logic/vals/projects.js';
import { valsTickets } from '~/logic/vals/tickets.js';
import { valsShell } from '~/logic/vals/shell.js';
import { valsHome } from '~/logic/vals/home.js';
import { valsPanels } from '~/logic/vals/panels.js';

// Hub logic. The pieces live in logic/:
//   state/    initial state, grouped by area
//   actions/  methods mixed into HubLogic
//   views/    renderVals() steps, run in order; each returns the names later steps need
//   vals/     the values and handlers the templates bind to
//   projects.js  milestone, objective, task and sprint rules
export class HubLogic extends DCLogic {
  constructor(props) {
    super(props);
    this.state = Object.assign({}, coreState(), awardsState(), calendarState(), commsState(), adminState(), profileState(), filesState(), projectsState(), callsState(), ticketsState(), wallState(), settingsState(), notesState(), forumsState(), onboardingState());
    // Nothing is made up: everything below arrives from the server after sign-in.
    this.dir = [];
    this.org = { sectors: [], branches: [], units: [], departments: [], mine: {} };
    this.baseLogs = [];
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
      if (patch.section !== undefined && patch.section !== this.state.section &&
          this.isMobileViewport() && ['dm', 'forum'].includes(this.state.panelKind) &&
          !['closed', 'min'].includes(this.state.mode)) {
        patch.mode = 'min';
      }
    }
    super.setState(patch, done);
    // A ringtone only plays while a call is incoming and not yet answered; a ringback only until the call connects.
    if (patch && ('incoming' in patch || 'call' in patch)) {
      const st = this.state;
      if (!st.incoming || st.call) stopRing();
      if (st.call && st.call.connected) stopRingback();
    }
  }
  renderVals() {
    const ctx = {};
    const prof = this._prof;   // development: { stepName: ms } filled in when set
    for (const step of VIEW_STEPS) { const t0 = prof ? performance.now() : 0; Object.assign(ctx, step.call(this, ctx)); if (prof) prof[step.name] = (prof[step.name] || 0) + performance.now() - t0; }
    const out = {};
    for (const part of VAL_PARTS) { const t0 = prof ? performance.now() : 0; Object.assign(out, part.call(this, ctx)); if (prof) prof[part.name] = (prof[part.name] || 0) + performance.now() - t0; }
    return out;
  }
}

Object.assign(HubLogic.prototype, messagingActions, projectsActions, callsActions, uiActions, forumActions, onboardingActions, peopleAdminActions, ticketCatActions, noteActions);

const VIEW_STEPS = [buildAccess, buildReporting, buildForums, buildNav, buildWall, buildTickets, buildForumExtras, buildProjects, buildPeople, buildOnboarding, buildSecurity, buildRoles, buildTableHelpers, buildProjectTables, buildCalendar, buildProjectStandup, buildOrg, buildCatMgr, buildComms, buildAdminTables, buildNotes, buildReportingTab, buildSettings, buildAwards, buildHomeWidgets, buildProfile, buildWidgetSettings, buildCallState];
const VAL_PARTS = [valsAccess, valsPages, valsProjects, valsTickets, valsShell, valsHome, valsWall, valsPanels];
