<!-- One comment and, indented under a thread line, every reply to it. Replies can be replied to again. -->
<template>
<div :style="node.depth > 0 ? `margin-left: ${node.depth > 4 ? 0 : 14}px; padding-left: 10px; border-left: 2px solid #2A3832` : ''">
<div style="display: flex; gap: 8px; align-items: flex-start; padding-top: 8px">
<div aria-hidden="true" :style="`width: ${node.depth ? 26 : 32}px; height: ${node.depth ? 26 : 32}px; flex: none; border-radius: 50%; background: ${node.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ node.initials }}</div>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="padding: 8px 12px; border-radius: 4px 16px 16px 16px; background: #1F2C26; min-width: 0"><strong style="font-size: 13px">{{ node.who }}</strong><span v-if="node.time" style="font-size: 11px; color: #9DB0A6"> · {{ node.time }}</span><div style="font-size: 14px; line-height: 1.4; overflow-wrap: anywhere">{{ node.text }}</div></div>
<div style="display: flex; gap: 12px; align-items: center; padding-left: 6px; font-size: 12px">
<button type="button" @click="node.replying ? node.cancel() : node.startReply()" :aria-expanded="node.replying" style="padding: 0; border: 0; background: transparent; color: #B8F36B; cursor: pointer; font-size: 12px; font-weight: 600">Reply</button>
<button v-if="node.replies.length" type="button" @click="collapsed = !collapsed" :aria-expanded="!collapsed" style="padding: 0; border: 0; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 12px">{{ collapsed ? 'Show' : 'Hide' }} {{ node.replies.length }} {{ node.replies.length === 1 ? 'reply' : 'replies' }}</button>
</div>
<CommentBox v-if="node.replying" :placeholder="`Reply to ${node.who}…`" send-label="Reply" cancellable autofocus @send="node.send" @cancel="node.cancel"/>
</div>
</div>
<template v-if="!collapsed"><template v-for="r in node.replies" :key="r.key"><CommentNode :node="r"/></template></template>
</div>
</template>

<script>
export default { name: 'CommentNode', props: { node: { type: Object, required: true } }, data() { return { collapsed: false }; } };
</script>
