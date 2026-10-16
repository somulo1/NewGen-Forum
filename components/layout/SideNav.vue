<!-- Sidebar: main pages, admin, messages, organization, forums, profile -->
<template><nav aria-label="Main" :class="navClass" @click="navClick" style="flex: 0 0 240px; overflow-y: auto; box-sizing: border-box; padding: 16px 12px; border-right: 1px solid #24322C; background: #121B17; display: flex; flex-direction: column; gap: 4px">
<label style="position: relative; display: block; margin: 0 0 8px"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Search the sidebar</span><input data-keep="1" type="search" :value="navSearch.q" @input="navSearch.onInput" placeholder="Search pages, people, forums…" aria-label="Search the sidebar" style="width: 100%; box-sizing: border-box; height: 38px; padding: 0 12px; border-radius: 19px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"></label>
<template v-if="navSearch.active">
<div role="list" aria-label="Search results" style="display: flex; flex-direction: column; gap: 2px">
<template v-for="(r, r__i) in navSearch.results" :key="r__i"><button @click="r.run" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 4px 10px; border: 0; border-radius: 10px; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; font-family: inherit"><span :style="`width: 30px; height: 30px; flex: none; border-radius: 9px; background: #22302A; color: ${r.color}; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ r.initials }}</span><span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ r.label }}</span><span style="font-size: 11px; color: #9DB0A6">{{ r.group }}<template v-if="r.sub"> · {{ r.sub }}</template></span></span></button></template>
<template v-if="navSearch.none"><span style="padding: 12px; font-size: 13px; color: #9DB0A6">Nothing matches "{{ navSearch.q }}".</span></template>
</div>
</template>
<template v-if="!navSearch.active">
<button @click="go.home" :style="nav.home"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M3 11l9-7 9 7v9a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z"></path></svg>Home</button>
<button @click="go.feed" :style="nav.feed"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="4" y="3" width="16" height="18" rx="2"></rect><path d="M8 8h8M8 12h8M8 16h5"></path></svg>Community wall</button>
<template v-if="has.sessions"><button @click="go.meetings" :style="nav.meetings"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg>Meetings &amp; standups</button></template>
<template v-if="has.calendar"><button @click="go.calendar" :style="nav.calendar"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="5" width="18" height="16" rx="2"></rect><path d="M3 10h18M8 3v4M16 3v4"></path></svg>Calendar</button></template>
<template v-if="has.comms"><button @click="go.comms" :style="nav.comms"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M4 10v4h3l6 5V5L7 10z"></path><path d="M17 9a4 4 0 0 1 0 6"></path></svg>Communications</button></template>
<template v-if="has.tickets"><button @click="go.tickets" :style="nav.tickets"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M3 7h18v4a2 2 0 0 0 0 4v4H3v-4a2 2 0 0 0 0-4z"></path></svg>My tickets</button></template>
<template v-if="ticketDeskOk"><button @click="go.ticketdesk" :style="nav.ticketdesk"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M3 7h18v10H3zM3 12h18M8 7v10"></path></svg>Ticket desk</button></template>
<template v-if="has.projects"><button @click="go.projects" :style="nav.projects"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M9 4v16M15 4v16M5.5 8h1.5M11.5 8h1.5M17.5 8h1"></path></svg>Projects</button></template>
<template v-if="has.files"><button @click="go.files" :style="nav.files"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M3 6h6l2 2h10v11H3z"></path></svg>Files</button></template>
<template v-if="has.secureshare"><button @click="go.share" :style="nav.share"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="8" cy="15" r="4"></circle><path d="M11 12l9-9M17 6l3 3"></path></svg>Secure share</button></template>
<button @click="go.people" :style="nav.people"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="9" cy="8" r="3.5"></circle><path d="M2 20c0-3.5 3-6 7-6s7 2.5 7 6M16 4.5a3.5 3.5 0 0 1 0 7M18 14c2.5.6 4 2.8 4 6"></path></svg>People</button>
<button @click="go.approvals" :style="nav.approvals"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3 8-8"></path><path d="M20 12v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h9"></path></svg><span style="flex: 1">Approvals</span><template v-if="hasPendingApprovals"><span style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #FFB86B; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ pendingApprovals }}</span></template></button>
<template v-if="canSeeAdmin">
<span style="margin: 14px 12px 4px; font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase; color: #9DB0A6">Admin</span>
</template>
<template v-if="canSecurity"><button @click="go.security" :style="nav.security"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 3l8 3v6c0 4.5-3.4 8.3-8 9-4.6-.7-8-4.5-8-9V6z"></path><path d="M9 12l2 2 4-4"></path></svg><span style="flex: 1">Security &amp; logs</span><template v-if="hasOpenAlerts"><span style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #FF9B8A; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ openAlertCount }}</span></template></button></template>
<template v-if="canRoles"><button @click="go.roles" :style="nav.roles"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="4" y="10" width="16" height="11" rx="2"></rect><path d="M8 10V7a4 4 0 0 1 8 0v3"></path></svg>Roles &amp; rights</button></template>

<template v-if="has.chat">
<div style="display: flex; align-items: center; justify-content: space-between; gap: 6px; margin: 16px 4px 6px 2px">
<button data-keep="1" @click="toggleDms" :aria-expanded="dmsOpen" style="flex: 1; display: flex; align-items: center; gap: 8px; min-height: 36px; padding: 0 8px; border: 0; border-radius: 8px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" :style="dmsChevron"><path d="M6 9l6 6 6-6"></path></svg>Messages<span style="margin-left: auto; padding: 1px 7px; border-radius: 9px; background: #22302A; color: #C9D6CF; font-size: 11px; letter-spacing: 0">{{ dmUnreadLabel }}</span></button>
<button aria-label="New private message" data-keep="1" @click="toggleNewDm" style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg></button>
</div>
<template v-if="newDmOpen">
<div style="margin: 0 4px 8px; padding: 10px; border-radius: 12px; background: #1A2520; border: 1px solid #2A3832; display: flex; flex-direction: column; gap: 6px">
<input type="search" :value="dmq" @input="onDmq" placeholder="Search anyone…" aria-label="Search people to message" style="height: 34px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px">
<template v-for="(dmm, dmm__i) in dmMatches" :key="dmm__i"><button @click="dmm.open" style="display: flex; align-items: center; gap: 8px; min-height: 38px; padding: 0 6px; border: 0; border-radius: 8px; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left"><span :style="`width: 26px; height: 26px; flex: none; border-radius: 50%; background: ${dmm.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 10px`">{{ dmm.initials }}</span><span style="flex: 1; min-width: 0; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ dmm.name }}</span></button></template>
</div>
</template>
<template v-if="dmsOpen">
<template v-for="(dl, dl__i) in dmList" :key="dl__i">
<button @click="dl.open" :style="dl.style">
<span :style="`position: relative; width: 32px; height: 32px; flex: none; border-radius: 50%; background: ${dl.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ dl.initials }}<span :style="`position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: ${dl.dot}; border: 2px solid #121B17`"></span></span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ dl.name }}</span><span style="font-size: 11px; color: #9DB0A6; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ dl.last }}</span></span>
<template v-if="dl.hasUnread"><span style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #C9B6FF; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ dl.unread }}</span></template>
</button>
</template>
</template>
</template>

<template v-if="myUnits.length">
<button data-keep="1" @click="toggleOrg" :aria-expanded="orgOpen" style="margin: 16px 4px 4px 2px; display: flex; align-items: center; gap: 8px; min-height: 36px; padding: 0 8px; border: 0; border-radius: 8px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" :style="orgChevron"><path d="M6 9l6 6 6-6"></path></svg>Organization</button>
<template v-if="orgOpen">
<template v-for="(mu, mu__i) in myUnits" :key="mu__i"><button @click="mu.open" :style="mu.style"><span :style="`width: 30px; height: 30px; flex: none; border-radius: 9px; background: #22302A; color: ${mu.color}; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 10px`">{{ mu.short }}</span><span style="flex: 1; min-width: 0; display: flex; flex-direction: column; text-align: left"><span style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ mu.name }}</span><span style="font-size: 11px; color: #9DB0A6">My {{ mu.typeLower }}</span></span></button></template>
</template>
</template>
<div style="display: flex; align-items: center; justify-content: space-between; gap: 6px; margin: 16px 4px 6px 2px">
<button data-keep="1" @click="toggleForums" :aria-expanded="forumsOpen" style="flex: 1; display: flex; align-items: center; gap: 8px; min-height: 36px; padding: 0 8px; border: 0; border-radius: 8px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" :style="forumsChevron"><path d="M6 9l6 6 6-6"></path></svg>Forums<span style="margin-left: auto; padding: 1px 7px; border-radius: 9px; background: #22302A; color: #C9D6CF; font-size: 11px; letter-spacing: 0">{{ forumCount }}</span></button>
<button aria-label="Create a forum" data-keep="1" @click="toggleCreate" style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg></button>
</div>

<template v-if="createOpen">
<div style="margin: 4px 4px 10px; padding: 12px; border-radius: 12px; background: #1A2520; border: 1px solid #2A3832; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; gap: 10px; align-items: center">
<button aria-label="Upload forum logo" style="width: 44px; height: 44px; border-radius: 12px; border: 1px dashed #4A5C54; background: transparent; color: #9DB0A6; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg></button>
<label style="flex: 1; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #9DB0A6">Forum name<input type="text" placeholder="e.g. Finance Team" style="height: 34px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label>
</div>
<label style="display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #9DB0A6">Invite members<input type="text" placeholder="Names or emails" style="height: 34px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label>
<button data-keep="1" @click="toggleCreate" style="height: 38px; border-radius: 9px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Create forum</button>
</div>
</template>

<template v-if="forumsOpen">
<template v-for="(f, f__i) in forums" :key="f__i">
<button @click="f.open" :style="f.rowStyle">
<span :style="`width: 30px; height: 30px; flex: none; border-radius: 9px; background: ${f.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 12px`">{{ f.initials }}</span>
<span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</span>
<template v-if="f.hasUnread"><span style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #B8F36B; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ f.unread }}</span></template>
</button>
</template>
</template>

<button @click="go.profile" :style="nav.profileCard">
<span style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; background: #C2185B; color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600; overflow: hidden">
<template v-if="me.hasPhoto"><img :src="me.photo" alt="" style="width: 100%; height: 100%; object-fit: cover"></template>
<template v-if="me.noPhoto">S</template>
</span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column; text-align: left"><strong style="font-size: 13px">My profile</strong><span :style="`font-size: 11px; color: ${me.statusColor}`">● {{ me.status }}</span></span>
</button>
</template>
</nav>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'SideNav',
  ...hubBindings(["navSearch","canRoles","canSecurity","canSeeAdmin","createOpen","dmList","dmMatches","dmUnreadLabel","dmq","dmsChevron","dmsOpen","forumCount","forums","forumsChevron","forumsOpen","go","has","hasOpenAlerts","hasPendingApprovals","me","myUnits","nav","navClass","navClick","newDmOpen","onDmq","openAlertCount","orgChevron","orgOpen","pendingApprovals","ticketDeskOk","toggleCreate","toggleDms","toggleForums","toggleNewDm","toggleOrg"])
};
</script>
