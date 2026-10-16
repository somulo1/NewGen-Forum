<!-- App shell: header, sidebar, page slot, docked panel and every overlay. -->
<template>
<div class="dc-root" style="display: contents">
<div :class="themeClass" :style="`height: 100vh; overflow: hidden; background: #0F1714; color: #E8EFEA; font-family: 'DM Sans', system-ui, sans-serif; font-size: 14px; display: flex; flex-direction: column; user-select: ${selectMode}; filter: ${themeFilter}`">

<AppHeader />
<PreviewBanner />

<div style="flex: 1; min-height: 0; display: flex; align-items: stretch">

<NavBackdrop />
<SideNav />

<div style="flex: 1 1 auto; min-width: 0; min-height: 0; display: flex; align-items: stretch">
<div style="position: relative; flex: 1 1 auto; min-width: 0; min-height: 0; display: flex; align-items: stretch">
<main class="hub-main" style="flex: 1 1 auto; min-width: 0; overflow-y: auto; box-sizing: border-box; padding: 24px clamp(16px, 3vw, 40px) 120px">
<slot />
</main>
<CallScreen variant="main" />
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

<WidgetSettingsDrawer />

<CallMini />

<AppToast />

<WidgetGallery />

<MobileTabBar />

<PanelBubble />

</div>
</div>

</template>

<script>
import { hubBindings, useHubShell } from '~/composables/useHub';

export default {
  name: 'DefaultLayout',
  ...hubBindings(["selectMode","themeClass","themeFilter"]),
  setup() { useHubShell(); }
};
</script>
