<!-- My profile -->
<template><template v-if="show.profile">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">

<section aria-label="Profile header" style="border-radius: 22px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<div :style="me.coverStyle">
<svg aria-hidden="true" width="100%" height="100%" viewBox="0 0 800 180" preserveAspectRatio="xMidYMid slice" style="position: absolute; inset: 0"><g fill="none" stroke="rgba(255,255,255,0.09)" stroke-width="1.5"><circle cx="660" cy="40" r="60"></circle><circle cx="660" cy="40" r="110"></circle><circle cx="660" cy="40" r="160"></circle><circle cx="660" cy="40" r="210"></circle><path d="M0 150 L120 110 L220 135 L340 80 L460 120 L560 70"></path></g></svg>
<template v-if="me.editing">
<div style="position: absolute; top: 14px; right: 14px; display: flex; gap: 6px; align-items: center; padding: 6px 8px; border-radius: 20px; background: rgba(10,16,13,0.7)"><span style="font-size: 12px; padding: 0 4px">Cover</span><template v-for="(cv, cv__i) in me.covers" :key="cv__i"><button :aria-label="cv.label" @click="cv.pick" :style="cv.style"></button></template></div>
</template>
</div>
<div style="padding: 0 24px 20px; display: flex; flex-wrap: wrap; gap: 18px; align-items: flex-end; margin-top: -54px; position: relative">
<div style="position: relative; flex: none">
<span :style="`width: 124px; height: 124px; border-radius: 50%; border: 5px solid var(--k-16201C,#16201C); background: ${me.color}; color: ${me.isOther ? '#0F1714' : '#fff'}; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-size: 46px; font-weight: 700; overflow: hidden; box-sizing: border-box`">
<template v-if="me.hasPhoto"><img :src="me.photo" alt="Profile photo" style="width: 100%; height: 100%; object-fit: cover"></template>
<template v-if="me.noPhoto">{{ me.initials }}</template>
</span>
<span aria-hidden="true" :style="`position: absolute; right: 8px; bottom: 8px; width: 22px; height: 22px; border-radius: 50%; background: ${me.statusColor}; border: 4px solid var(--k-16201C,#16201C)`"></span>
<template v-if="me.editing"><label style="position: absolute; left: 50%; bottom: -6px; transform: translateX(-50%); height: 30px; padding: 0 12px; border-radius: 15px; background: var(--k-E8EFEA,#E8EFEA); color: var(--k-0F1714,#0F1714); font-size: 12px; font-weight: 600; white-space: nowrap; display: flex; align-items: center; cursor: pointer"><input type="file" accept="image/*" aria-label="Change photo" @change="onPhoto" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer">Change photo</label></template>
</div>
<div style="flex: 1 1 300px; min-width: 0; display: flex; flex-direction: column; gap: 6px; padding-bottom: 2px">
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">{{ me.name }}</h1>
<span style="font-size: 14px; color: var(--k-C9D6CF,#C9D6CF)">{{ me.role }} · {{ me.dept }}</span>
<template v-if="me.isOther"><span style="align-self: flex-start; height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); display: flex; align-items: center; gap: 8px; font-size: 13px"><span :style="`width: 9px; height: 9px; border-radius: 50%; background: ${me.statusColor}`"></span>{{ me.status }}</span></template>
<div v-if="!me.isOther" style="position: relative; align-self: flex-start">
<button @click="toggleStatusMenu" aria-haspopup="menu" style="height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; gap: 8px; font-size: 13px"><span :style="`width: 9px; height: 9px; border-radius: 50%; background: ${me.statusColor}`"></span>{{ me.status }}<svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"></path></svg></button>
<template v-if="me.statusMenu">
<div role="menu" style="position: absolute; top: 38px; left: 0; z-index: 20; width: 220px; padding: 6px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 14px 30px rgba(0,0,0,0.45)">
<template v-for="(so, so__i) in me.statuses" :key="so__i"><button role="menuitem" @click="so.pick" style="display: flex; align-items: center; gap: 10px; width: 100%; min-height: 40px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 14px; text-align: left"><span :style="`width: 9px; height: 9px; border-radius: 50%; background: ${so.color}`"></span>{{ so.label }}</button></template>
</div>
</template>
</div>
</div>
<div style="display: flex; gap: 8px; flex-wrap: wrap">
<template v-if="me.isOther"><button @click="profileBack" style="height: 42px; padding: 0 20px; border-radius: 21px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Back to people</button></template>
<template v-if="me.notEditing && !me.isOther">
<button @click="startEditProfile" style="height: 42px; padding: 0 20px; border-radius: 21px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Edit profile</button>
<button @click="shareProfile" style="height: 42px; padding: 0 18px; border-radius: 21px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Share profile</button>
</template>
<template v-if="me.editing">
<button @click="saveProfile" style="height: 42px; padding: 0 20px; border-radius: 21px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save changes</button>
<button @click="cancelProfile" style="height: 42px; padding: 0 18px; border-radius: 21px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Done</button>
</template>
</div>
</div>
<div v-if="!me.isOther" style="padding: 0 24px 20px; display: flex; flex-wrap: wrap; gap: 12px; align-items: center">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Profile {{ me.complete }}% complete</span>
<div style="flex: 1 1 200px; max-width: 320px; height: 6px; border-radius: 3px; background: var(--k-22302A,#22302A); overflow: hidden"><div :style="me.completeBar"></div></div>
<span style="font-size: 12px; color: var(--k-C9D6CF,#C9D6CF)">{{ me.completeTip }}</span>
</div>
</section>
<StorageCard v-if="!me.isOther" />

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 12px">
<template v-for="(fc, fc__i) in me.facts" :key="fc__i">
<div style="padding: 14px 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 4px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ fc.label }}</span><strong style="font-size: 14px; font-weight: 600">{{ fc.value }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ fc.sub }}</span></div>
</template>
</div>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 12px">
<template v-for="(st, st__i) in me.stats" :key="st__i">
<div style="padding: 14px 16px; border-radius: 16px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 28px; font-weight: 700; color: ${st.color}`">{{ st.n }}</span><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ st.label }}</span></div>
</template>
</div>

<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
<div style="flex: 999 1 480px; min-width: 0; display: flex; flex-direction: column; gap: 16px">

<div style="padding: 20px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<strong style="font-size: 15px">{{ me.aboutTitle }}</strong>
<template v-if="me.notEditing"><p style="margin: 0; font-size: 15px; line-height: 1.6; color: var(--k-C9D6CF,#C9D6CF); white-space: pre-line">{{ me.bio }}</p></template>
<template v-if="me.editing"><label><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">About me</span><textarea rows="4" :value="me.bio" @input="onBio" style="width: 100%; box-sizing: border-box; padding: 12px; border-radius: 12px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 15px; line-height: 1.5; resize: vertical"></textarea></label></template>
</div>

<div style="padding: 20px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 15px">Skills &amp; expertise</strong>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-for="(sk, sk__i) in me.skills" :key="sk__i">
<span style="height: 34px; padding: 0 14px; border-radius: 17px; background: var(--k-1F2C26,#1F2C26); border: 1px solid var(--k-2A3832,#2A3832); font-size: 13px; display: flex; align-items: center; gap: 6px">{{ sk.name }}<template v-if="sk.editing"><button :aria-label="`Remove ${sk.name}`" @click="sk.remove" style="width: 22px; height: 22px; margin-right: -6px; border-radius: 50%; border: 0; background: var(--k-2C3D34,#2C3D34); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></template></span>
</template>
</div>
<template v-if="me.editing"><div style="display: flex; gap: 8px"><label style="flex: 1; min-width: 0"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Add a skill</span><input type="text" :value="me.newSkill" @input="onNewSkill" @keydown="newSkillKey" placeholder="Add a skill and press Enter" style="width: 100%; box-sizing: border-box; height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label><button @click="addSkill" style="height: 38px; padding: 0 14px; border-radius: 10px; border: 1px dashed var(--k-4A5C54,#4A5C54); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Add</button></div></template>
</div>

<div style="padding: 20px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 15px">Working on</strong>
<span v-if="!me.projects.length" style="font-size: 14px; color: var(--k-9DB0A6,#9DB0A6)">No active projects right now.</span>
<template v-for="(mp, mp__i) in me.projects" :key="mp__i">
<div style="display: flex; gap: 12px; align-items: center; padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17)">
<span :style="`width: 38px; height: 38px; flex: none; border-radius: 11px; background: ${mp.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 13px`">{{ mp.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6px">
<div style="display: flex; gap: 8px; align-items: baseline; flex-wrap: wrap"><strong style="font-size: 14px">{{ mp.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ mp.myRole }} · <span :style="`color: ${mp.statusColor}`">{{ mp.status }}</span></span></div>
<div style="display: flex; align-items: center; gap: 8px"><div style="flex: 1; height: 6px; border-radius: 3px; background: var(--k-22302A,#22302A); overflow: hidden"><div :style="mp.barStyle"></div></div><span style="font-size: 11px; font-weight: 600">{{ mp.progress }}%</span></div>
</div>
<button @click="mp.open" style="flex: none; height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; font-weight: 600">View</button>
</div>
</template>
</div>

<div style="padding: 20px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px">
<strong style="font-size: 15px">Recent activity</strong>
<div tabindex="0" aria-label="Recent activity table" style="max-height: 300px; overflow: auto; border-radius: 12px; border: 1px solid var(--k-24322C,#24322C)">
<table style="width: 100%; min-width: 460px; border-collapse: separate; border-spacing: 0; font-size: 14px">
<thead><tr>
<th style="position: sticky; top: 0; z-index: 1; padding: 10px 14px; text-align: left; font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822); white-space: nowrap">Type</th>
<th style="position: sticky; top: 0; z-index: 1; padding: 10px 14px; text-align: left; font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822)">Activity</th>
<th style="position: sticky; top: 0; z-index: 1; padding: 10px 14px; text-align: left; font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822); white-space: nowrap">When</th>
</tr></thead>
<tbody>
<template v-for="(ma, ma__i) in me.activity" :key="ma__i">
<tr>
<td style="padding: 10px 14px; border-top: 1px solid var(--k-24322C,#24322C); white-space: nowrap"><span :style="`height: 24px; padding: 0 10px; border-radius: 12px; background: var(--k-22302A,#22302A); color: ${ma.color}; display: inline-flex; align-items: center; font-size: 12px; font-weight: 600`">{{ ma.kind }}</span></td>
<td style="padding: 10px 14px; border-top: 1px solid var(--k-24322C,#24322C); line-height: 1.45">{{ ma.text }}</td>
<td style="padding: 10px 14px; border-top: 1px solid var(--k-24322C,#24322C); color: var(--k-9DB0A6,#9DB0A6); white-space: nowrap">{{ ma.time }}</td>
</tr>
</template>
</tbody>
</table>
</div>
</div>
</div>

<aside aria-label="Profile details" style="flex: 1 1 300px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Contact</span>
<template v-for="(ct, ct__i) in me.contact" :key="ct__i"><div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ ct.label }}</span><span :style="`font-size: 14px; color: ${ct.color}`">{{ ct.value }}</span></div></template>
<template v-if="me.editing"><label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Phone number<input type="tel" :value="me.phone" @input="onPhone" placeholder="Add your work phone" style="height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label></template>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; margin-top: 4px"><button @click="copyEmail" style="height: 38px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Copy email</button><button @click="shareProfile" style="height: 38px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Share card</button></div>
</div>

<section v-if="!me.isOther" aria-label="Change password" style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<div><strong style="font-size: 15px">Change password</strong><p style="margin: 4px 0 0; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Choose a new password for your account.</p></div>
<form @submit.prevent="changePassword" style="display: flex; flex-direction: column; gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Current password<input v-model="currentPassword" type="password" autocomplete="current-password" required style="height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">New password<input v-model="newPassword" type="password" autocomplete="new-password" minlength="8" required style="height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Confirm new password<input v-model="confirmPassword" type="password" autocomplete="new-password" minlength="8" required style="height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<p v-if="passwordMessage" :role="passwordError ? 'alert' : 'status'" :style="`margin: 0; font-size: 13px; color: ${passwordError ? 'var(--k-FF9B8A,#FF9B8A)' : 'var(--k-B8F36B,#B8F36B)'}`">{{ passwordMessage }}</p>
<button type="submit" :disabled="passwordBusy" style="height: 40px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ passwordBusy ? 'Updating…' : 'Update password' }}</button>
</form>
</section>

<div style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Availability</span>
<div style="display: flex; align-items: baseline; gap: 10px"><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 600">{{ me.localTime }}</span><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">local time</span></div>
<span style="font-size: 14px; color: var(--k-C9D6CF,#C9D6CF)">Works Monday to Friday, {{ me.hours }}</span>
<span :style="`font-size: 13px; color: ${me.statusColor}`">● {{ me.statusNote }}</span>
</div>

<template v-if="me.hasBadges">
<div style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Badges &amp; awards</span>
<div tabindex="0" aria-label="Badges and awards, scrollable" style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px; max-height: 214px; overflow-y: auto; padding-right: 4px">
<template v-for="(bd, bd__i) in me.badges" :key="bd__i">
<div style="padding: 12px 6px; border-radius: 12px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; align-items: center; gap: 6px; text-align: center">
<span :style="`width: 44px; height: 44px; border-radius: 50%; background: ${bd.bg}; color: ${bd.color}; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><circle cx="12" cy="9" r="6"></circle><path d="M8.5 14l-1.5 7 5-3 5 3-1.5-7"></path></svg></span>
<span style="font-size: 12px; font-weight: 600; line-height: 1.3">{{ bd.name }}</span>
<span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ bd.when }}</span>
</div>
</template>
</div>
</div>
</template>

<div style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Team</span>
<div style="display: flex; gap: 10px; align-items: center; padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17)"><span aria-hidden="true" style="width: 34px; height: 34px; border-radius: 50%; background: #6EE7C8; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px">HR</span><div style="display: flex; flex-direction: column"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Reports to</span><strong style="font-size: 14px; font-weight: 600">{{ me.reportsTo }}</strong></div></div>
</div>

<div style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">{{ me.forumsTitle }}</span>
<template v-for="(mf, mf__i) in me.forums" :key="mf__i"><button @click="mf.open" style="display: flex; align-items: center; gap: 10px; min-height: 40px; padding: 4px 8px; border: 0; border-radius: 10px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left"><span :style="`width: 28px; height: 28px; border-radius: 8px; background: ${mf.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 10px`">{{ mf.initials }}</span><span style="flex: 1; font-size: 14px">{{ mf.name }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ mf.myRole }}</span></button></template>
</div>
</aside>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ProfileSection',
  ...hubBindings(["addSkill","copyEmail","cancelProfile","me","newSkillKey","onBio","onNewSkill","onPhone","onPhoto","profileBack","saveProfile","shareProfile","show","startEditProfile","toggleStatusMenu"]),
  data() {
    return { currentPassword: '', newPassword: '', confirmPassword: '', passwordMessage: '', passwordError: false, passwordBusy: false };
  },
  methods: {
    async changePassword() {
      this.passwordMessage = '';
      this.passwordError = false;
      if (this.newPassword.length < 8) {
        this.passwordError = true;
        this.passwordMessage = 'Use a password of at least 8 characters.';
        return;
      }
      if (this.newPassword !== this.confirmPassword) {
        this.passwordError = true;
        this.passwordMessage = 'The new passwords do not match.';
        return;
      }
      const session = useAuth().user();
      const email = session && session.email;
      if (!email) {
        this.passwordError = true;
        this.passwordMessage = 'Sign in to change your password.';
        return;
      }
      this.passwordBusy = true;
      try {
        if (useBackend().signedIn()) {
          await useBackend().post('/auth/password', { current_password: this.currentPassword, new_password: this.newPassword });
          this.currentPassword = '';
          this.newPassword = '';
          this.confirmPassword = '';
          this.passwordMessage = 'Password updated successfully.';
          return;
        }
        const accounts = useAccounts();
        if (!accounts.get(email)) {
          this.passwordError = true;
          this.passwordMessage = 'No local account exists for this sign-in. Contact your administrator.';
          return;
        }
        if (!await accounts.verifyPassword(email, this.currentPassword)) {
          this.passwordError = true;
          this.passwordMessage = 'Your current password is incorrect.';
          return;
        }
        await accounts.setPassword(email, this.newPassword);
        this.currentPassword = '';
        this.newPassword = '';
        this.confirmPassword = '';
        this.passwordMessage = 'Password updated successfully.';
      } catch (error) {
        this.passwordError = true;
        this.passwordMessage = 'Could not update your password. ' + (error instanceof Error ? error.message : String(error));
      } finally {
        this.passwordBusy = false;
      }
    }
  }
};
</script>
