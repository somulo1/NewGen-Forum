<!-- Memo / announcement / newsletter composer -->
<template><template v-if="cmOpen">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 66; background: rgba(5,10,8,0.65); display: flex; align-items: center; justify-content: center; padding: 16px; box-sizing: border-box">
<div role="dialog" :aria-label="cm.heading" style="width: 100%; max-width: 660px; max-height: 92vh; display: flex; flex-direction: column; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); box-shadow: 0 24px 60px rgba(0,0,0,0.5); overflow: hidden">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 16px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)"><h2 style="margin: 0; flex: 1; font-family: 'Space Grotesk', sans-serif; font-size: 20px">{{ cm.heading }}</h2><button aria-label="Close" @click="cm.close" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(ct, ct__i) in cm.types" :key="ct__i"><button @click="ct.pick" :style="ct.style">{{ ct.label }}</button></template></div>
<span v-if="cm.typeHelp" style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ cm.typeHelp }}</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ cm.titleLabel }}<input type="text" :value="cm.title" @input="cm.setTitle" maxlength="140" placeholder="Give it a clear title" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"/></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Summary<input type="text" :value="cm.summary" @input="cm.setSummary" maxlength="200" placeholder="One line shown in lists" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"/><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Optional. People see this before they open it.</span></label>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 12px">
<template v-for="f in cm.fields" :key="f.key"><label :for="f.id" style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ f.label }}
<select v-if="f.isSelect" :id="f.id" :value="f.value" @change="f.onInput" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">Choose…</option><template v-for="o in f.options" :key="o"><option :value="o">{{ o }}</option></template></select>
<textarea v-else-if="f.isArea" :id="f.id" rows="3" :value="f.value" @input="f.onInput" style="padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; resize: vertical; font-family: inherit"></textarea>
<input v-else :id="f.id" :type="f.type" :value="f.value" @input="f.onInput" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"/>
<span v-if="f.help" style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.help }}</span></label></template>
</div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ cm.bodyLabel }}<textarea rows="6" :value="cm.body" @input="cm.setBody" placeholder="Write it here" style="padding: 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; resize: vertical; font-family: inherit"></textarea></label>
<div style="display: flex; flex-direction: column; gap: 8px; padding: 12px 14px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C)">
<label style="display: inline-flex; align-self: flex-start; align-items: center; gap: 8px; height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); cursor: pointer; font-size: 13px; font-weight: 600">Attach files<input type="file" multiple :accept="cm.accept" @change="cm.pickFiles" style="display: none"/></label>
<span v-if="cm.fileRule" style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ cm.fileRule }}</span>
<template v-for="(f, fi) in cm.files" :key="fi"><div style="display: flex; align-items: center; gap: 8px; font-size: 13px"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.size }}</span><button type="button" @click="f.remove" aria-label="Remove file" style="border: 0; background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer">Remove</button></div></template>
</div>
<div style="display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C)">
<strong style="font-size: 14px">Where should it go?</strong>
<template v-if="cm.wallAllowed"><label style="display: flex; gap: 10px; align-items: center; font-size: 14px; cursor: pointer"><input type="checkbox" :checked="cm.wall" @change="cm.toggleWall">Community wall</label></template>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">A forum<select :value="cm.forum" @change="cm.setForum" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">No forum</option><template v-for="(f, f__i) in forums" :key="f__i"><option :value="f.id">{{ f.name }}</option></template></select></label>
<template v-if="cm.hasRuleNote"><span style="font-size: 12px; color: var(--k-C9D6CF,#C9D6CF); line-height: 1.5; padding: 8px 10px; border-radius: 8px; background: var(--k-16201C,#16201C)">{{ cm.ruleNote }}</span></template>
<template v-for="(gr, gr__i) in cm.groups" :key="gr__i">
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ gr.label }}</span><div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(gi, gi__i) in gr.items" :key="gi__i"><button @click="gi.toggle" :aria-pressed="gi.on" :style="gi.style">{{ gi.name }}</button></template></div></div>
</template>
<template v-if="cm.noStructure"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Department, sector, branch and unit options appear here once your organization adds those widgets.</span></template>
</div>
<template v-if="cm.needsApproval"><span style="font-size: 13px; color: var(--k-FFB86B,#FFB86B)">{{ cm.approvalNote }}</span></template>
</div>
<div style="flex: none; display: flex; justify-content: flex-end; gap: 8px; padding: 14px 20px; border-top: 1px solid var(--k-24322C,#24322C)"><button @click="cm.close" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button><button @click="cm.publish" :style="cm.publishStyle">{{ cm.publishLabel }}</button></div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CommComposer',
  ...hubBindings(["cm","cmOpen","forums"])
};
</script>
