<!-- Grey blocks where a page's content will be, shown until the server has answered. Never a spinner: the shape tells the person
     what is coming, and nothing made up is shown in its place. -->
<template>
<div class="sk" role="status" aria-busy="true" aria-live="polite" style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<span class="sk-only">Loading…</span>
<div style="display: flex; flex-direction: column; gap: 10px"><span class="sk-b" style="width: min(320px, 60%); height: 30px"></span><span class="sk-b" style="width: min(520px, 90%); height: 14px"></span></div>

<template v-if="kind === 'home'">
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px">
<div v-for="n in 6" :key="n" class="sk-card" style="height: 210px; display: flex; flex-direction: column; gap: 12px"><span class="sk-b" style="width: 45%; height: 16px"></span><span class="sk-b" style="width: 100%; height: 12px"></span><span class="sk-b" style="width: 85%; height: 12px"></span><span class="sk-b" style="width: 92%; height: 12px"></span><span class="sk-b" style="width: 60%; height: 12px"></span></div>
</div>
</template>

<template v-else-if="kind === 'feed'">
<div style="display: flex; gap: 10px; overflow: hidden"><span v-for="n in 6" :key="n" class="sk-b" style="flex: none; width: 112px; height: 180px; border-radius: 14px"></span></div>
<div v-for="n in 3" :key="n" class="sk-card" style="display: flex; flex-direction: column; gap: 12px; max-width: 760px">
<div style="display: flex; align-items: center; gap: 12px"><span class="sk-b" style="width: 42px; height: 42px; border-radius: 50%"></span><div style="display: flex; flex-direction: column; gap: 6px"><span class="sk-b" style="width: 160px; height: 14px"></span><span class="sk-b" style="width: 100px; height: 11px"></span></div></div>
<span class="sk-b" style="width: 100%; height: 12px"></span><span class="sk-b" style="width: 80%; height: 12px"></span><span class="sk-b" style="width: 100%; height: 120px"></span>
</div>
</template>

<template v-else-if="kind === 'calendar'">
<div style="display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 6px"><span v-for="n in 35" :key="n" class="sk-b" style="height: 78px; border-radius: 10px"></span></div>
</template>

<template v-else-if="kind === 'cards'">
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 14px">
<div v-for="n in 8" :key="n" class="sk-card" style="height: 130px; display: flex; flex-direction: column; gap: 10px"><span class="sk-b" style="width: 40%; height: 16px"></span><span class="sk-b" style="width: 90%; height: 12px"></span><span class="sk-b" style="width: 70%; height: 12px"></span></div>
</div>
</template>

<template v-else-if="kind === 'form'">
<div class="sk-card" style="max-width: 760px; display: flex; flex-direction: column; gap: 16px">
<div v-for="n in 5" :key="n" style="display: flex; flex-direction: column; gap: 8px"><span class="sk-b" style="width: 140px; height: 12px"></span><span class="sk-b" style="width: 100%; height: 40px"></span></div>
</div>
</template>

<template v-else>
<div style="display: flex; gap: 8px"><span class="sk-b" style="width: 120px; height: 38px; border-radius: 19px"></span><span class="sk-b" style="width: 120px; height: 38px; border-radius: 19px"></span></div>
<div class="sk-card" style="padding: 0; overflow: hidden">
<div style="display: flex; gap: 16px; padding: 14px 16px; border-bottom: 1px solid var(--k-24322C,#24322C)"><span v-for="n in 5" :key="n" class="sk-b" style="flex: 1; height: 12px"></span></div>
<div v-for="n in 7" :key="n" style="display: flex; align-items: center; gap: 16px; padding: 16px; border-bottom: 1px solid var(--k-24322C,#24322C)"><span class="sk-b" style="width: 34px; height: 34px; border-radius: 50%; flex: none"></span><span class="sk-b" style="flex: 2; height: 13px"></span><span class="sk-b" style="flex: 1; height: 13px"></span><span class="sk-b" style="flex: 1; height: 13px"></span><span class="sk-b" style="width: 80px; height: 24px; border-radius: 12px"></span></div>
</div>
</template>
</div>
</template>

<script>
export default { name: 'PageSkeleton', props: { kind: { type: String, default: 'table' } } };
</script>

<style>
.sk-b { display: block; border-radius: 8px; background: linear-gradient(90deg, var(--k-1F2C26,#1F2C26) 0%, var(--k-2A3832,#2A3832) 45%, var(--k-1F2C26,#1F2C26) 90%); background-size: 220% 100%; animation: sk-shimmer 1.4s ease-in-out infinite; }
.sk-card { padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); box-sizing: border-box; }
.sk-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
@keyframes sk-shimmer { 0% { background-position: 120% 0; } 100% { background-position: -100% 0; } }
@media (prefers-reduced-motion: reduce) { .sk-b { animation: none; } }
</style>
