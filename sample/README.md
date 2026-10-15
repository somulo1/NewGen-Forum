# Hub — instructions for the coding agent

This Nuxt 3 project already contains the **complete, working Hub design**, split into normal Nuxt
folders. It is an exact port of the Claude Design prototype: every screen, colour, spacing value,
interaction and the mobile layout match one-to-one. Screenshots were pixel-compared to confirm it.

**Your job is to build on it, never to redesign it.**

## Hard rules
1. **Do not redesign, restyle or "improve" the UI.**
   - Do not add Tailwind, Vuetify, Nuxt UI, PrimeVue, shadcn or any other UI kit.
   - Do not change colours, fonts, radii, spacing or wording unless the user asks for that exact change.
2. **Treat the screenshots in `design-reference/` as the visual spec.** They cover desktop (1440×900)
   and mobile (390×844).
3. **Move markup and inline styles verbatim** when you refactor further, for example when you
   extract a widget card into its own component.
4. **After every change**, run `npm run dev`, then `npm run screens`, and compare `screens/` with
   `design-reference/`. If something changed that the user did not ask for, fix it first.
5. **Keep the responsive rules in `assets/hub.css`.** These are the `hub-*` classes, active at ≤900px.
   Every new screen must work at both 1440px and 390px wide.

## Folder map
```
app.vue                      NuxtLayout + NuxtPage
layouts/default.vue          App shell: header, sidebar, <main> page slot, docked panel, all overlays
pages/                       One page per section; each renders its section component
  index.vue (Home)  wall.vue  meetings.vue  calendar.vue  communications.vue  tickets.vue
  projects.vue  files.vue  secure-share.vue  people.vue  approvals.vue  security.vue  roles.vue
  profile.vue  org.vue (department/sector/branch/unit dashboard)  structure.vue
components/
  layout/    AppHeader, SideNav, NavBackdrop, MobileTabBar, PreviewBanner
  sections/  HomeSection, WallSection, TicketsSection (list + details), ProjectsSection (list + view),
             FilesSection, CalendarSection, MeetingsSection, CommsSection, PeopleSection,
             RolesSection, SecuritySection, ApprovalsSection, ProfileSection, OrgUnitSection, …
  panels/    ForumPanel (docked; full screen on mobile), ForumFull, PrivateChatFull, PanelBubble
  overlays/  AwardDialog, EventDialog, CommComposer, PersonDrawer, TaskDrawer,
             WidgetSettingsDrawer, WidgetGallery, CallScreen, AppToast
  ui/        DataTable: shared table with header filters (text/select/date range), pagination
             and cell kinds text, badge, person, avatars, check, select, bar, actions
logic/HubLogic.js            ALL state, sample data, business rules and actions (see below)
logic/routes.js              section ↔ URL map
composables/useHub.js        shared HubLogic instance, hubBindings(), URL ↔ section sync
utils/dcLogic.js             setState()/forceUpdate() adapter (React-style state on Vue reactivity)
assets/hub.css               global styles, light theme, responsive (mobile) rules
prototype-source/            original Claude Design files + generated single-file version
design-reference/            reference screenshots
scripts/                     convert/split/sync tools, screenshot script
```

## How data flows
- `HubLogic` holds `state`. Its `renderVals()` returns every value the templates use: lists,
  computed styles, labels and click handlers.
- Each component lists the names it uses in `hubBindings([...])`. Each name becomes a computed
  property, so templates read `tk.title`, `forums`, `openGallery` and so on directly.
- Actions are methods or closures in `HubLogic` that call `this.setState({...})`, for example
  `openDm`, `startCall`, `updTicket`, `addForumMsg` and `makeRequest`. Templates bind them with
  `@click="name"`.
- **To add a feature:**
  1. Add state in the `HubLogic` constructor.
  2. Compute what the template needs in `renderVals()` and return it.
  3. Use it in a component and add the name to that component's `hubBindings` list.

## Business rules to preserve
- **Widget gating:** a feature that is not added as a widget must not appear anywhere. That covers
  navigation, other pages, notifications, rights, the calendar and dashboards.
- **Rights:** rights matrix per role, plus acting-role preview. Changing a right creates an approval
  and shows as PENDING until it is approved.
- **Approvals:** flows with 0, 1, 2 or 3 levels and an approvals inbox.
- **Org structure:** Units sit under Branches, Branches under Sectors, Sectors under the Org.
  Departments attach to a branch, else a sector, else the org. A level exists only if its widget
  is added. Communication rules are set per level.
- **Tickets:**
  - TAT per category, with pause rules and notifications.
  - Escalation goes destination type → entity → a category that entity handles → a person assigned
    to that category.
- **Awards:** created with approval. Approved awards appear in the Home ribbon and as profile
  badges, and can be posted to the wall and forums with every awardee tagged.
- **Private chat:** uses the same docked, full-screen and minimized modes as forums. A call from a
  private chat is between the two people only.

## Recommended next steps (when asked to productionise)
1. **Move state into Pinia stores per module.** Suggested stores: widgets, forums, dms, tickets,
   projects, people, approvals, rights, org and calendar. Keep the field names. Do it gradually and
   compare screenshots after each module.
2. **Replace sample data with API calls** behind `composables/useApi.ts`. The backend is not chosen yet.
3. **Make simulated features real:** realtime messages, calls (WebRTC or a vendor SDK) and file uploads.

## When the design changes in Claude Design
1. Put the new `Main.dc.html` and `Table.dc.html` into `prototype-source/`.
2. Run `npm run sync`. This regenerates the layout, pages, components and logic, so commit first and
   re-apply any hand edits from `git diff`.

## Commands
- `npm install` then `npm run dev`: opens http://localhost:3000.
- `npx playwright install chromium` once, then `npm run screens`: screenshots every screen into `screens/`.
- `npm run generate`: static build in `.output/public`.
