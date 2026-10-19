<!-- App shell: header, sidebar, page slot, docked panel and every overlay. -->
<template>
<div class="dc-root" style="display: contents">
<div :class="themeClass" :style="`height: 100vh; overflow: hidden; background: var(--k-0F1714,#0F1714); color: var(--k-E8EFEA,#E8EFEA); font-family: 'DM Sans', system-ui, sans-serif; font-size: 14px; display: flex; flex-direction: column; user-select: ${selectMode}`">

<AppHeader />
<PreviewBanner />

<div style="flex: 1; min-height: 0; display: flex; align-items: stretch">

<NavBackdrop />
<SideNav />

<div style="flex: 1 1 auto; min-width: 0; min-height: 0; display: flex; align-items: stretch">
<div style="position: relative; flex: 1 1 auto; min-width: 0; min-height: 0; display: flex; align-items: stretch">
<main class="hub-main" style="flex: 1 1 auto; min-width: 0; overflow-y: auto; box-sizing: border-box; padding: 24px clamp(4px, 0.5vw, 8px) 120px clamp(16px, 3vw, 40px)">
<ErrorPage v-if="fatal" code="500" @leave="clearFatal" />
<ErrorPage v-else-if="denied" :code="denied.reason === 'off' ? '404' : '403'" :section="denied.label" :reason="denied.reason" />
<slot v-else />
</main>
<CallScreen variant="main" />
<ForumFull fit="page" />
<PrivateChatFull fit="page" />
</div>

<div class="hub-rightcol" style="flex: none; min-height: 0; display: flex; flex-direction: column">
<CallScreen variant="dock" />

<ForumPanel />
</div>
</div>
</div>

<PrivateChatFull />

<ForumFull />

<AwardDialog />

<EventDialog />

<CommComposer />

<PersonDrawer />

<TaskDrawer />

<MilestoneDialog />

<OnboardDialog />

<WidgetSettingsDrawer />

<IncomingCall />

<CallMini />

<AlertDialog />

<EtrDocument />

<FileViewer />

<AppToast />

<MobileTabBar />

<PanelBubble />

</div>
</div>

</template>

<script>
import { hubBindings, useHubShell } from '~/composables/useHub';

export default {
  name: 'DefaultLayout',
  ...hubBindings(["denied","selectMode","themeClass","themeFilter"]),
  setup() { useHubShell(); const fatal = useState('hubFatal', () => null); return { fatal, clearFatal: () => { fatal.value = null; } }; }
};
</script>
