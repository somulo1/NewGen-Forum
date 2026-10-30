<!-- One comment and, indented under a thread line, every reply to it. Replies can be replied to again. -->
<template>
<div :style="node.depth > 0 ? `margin-left: ${node.depth > 4 ? 0 : 14}px; padding-left: 10px; border-left: 2px solid var(--k-2A3832,#2A3832)` : ''">
<div style="display: flex; gap: 8px; align-items: flex-start; padding-top: 8px">
<UserAvatar :name="node.who" :initials="node.initials" :color="node.color" :size="node.depth ? 26 : 32" :font-size="11"/>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="padding: 8px 12px; border-radius: 4px 16px 16px 16px; background: var(--k-1F2C26,#1F2C26); min-width: 0"><strong style="font-size: 13px">{{ node.who }}</strong><span v-if="node.time" style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)"> · {{ node.time }}</span><div v-html="node.htmlText" style="font-size: 14px; line-height: 1.4; overflow-wrap: anywhere"></div></div>
<div style="display: flex; flex-wrap: wrap; gap: 4px 6px; align-items: center; padding-left: 2px; font-size: 12px">
<button v-if="node.canReact" type="button" @click="node.react('like')" :aria-pressed="node.iLiked" :aria-label="`Like. ${node.likes} ${node.likes === 1 ? 'like' : 'likes'}`" :style="chip(node.iLiked, '#B8F36B')"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" :fill="node.iLiked ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"><path d="M7 11v9H4v-9h3zm0 0l4-8c1.5 0 2.5 1 2.5 2.5V9h5.2a2 2 0 011.97 2.35l-1.1 6.5A2 2 0 0117.6 19H7"/></svg><span v-if="node.likes">{{ node.likes }}</span><span v-else>Like</span></button>
<button v-if="node.canReact" type="button" @click="node.react('dislike')" :aria-pressed="node.iDisliked" :aria-label="`Dislike. ${node.dislikes} ${node.dislikes === 1 ? 'dislike' : 'dislikes'}`" :style="chip(node.iDisliked, '#FF9B8A')"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" :fill="node.iDisliked ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round" style="transform: scaleY(-1)"><path d="M7 11v9H4v-9h3zm0 0l4-8c1.5 0 2.5 1 2.5 2.5V9h5.2a2 2 0 011.97 2.35l-1.1 6.5A2 2 0 0117.6 19H7"/></svg><span v-if="node.dislikes">{{ node.dislikes }}</span><span v-else>Dislike</span></button>
<button type="button" @click="node.replying ? node.cancel() : node.startReply()" :aria-expanded="node.replying" :style="chip(node.replying, '#7CC4FF')"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"><path d="M9 14L4 9l5-5M4 9h10a6 6 0 016 6v4"/></svg>Reply</button>
<button v-if="node.replies.length" type="button" @click="collapsed = !collapsed" :aria-expanded="!collapsed" style="height: 28px; padding: 0 8px; border: 0; border-radius: 14px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 12px">{{ collapsed ? 'Show' : 'Hide' }} {{ node.replies.length }} {{ node.replies.length === 1 ? 'reply' : 'replies' }}</button>
</div>
<CommentBox v-if="node.replying" :placeholder="`Reply to ${node.who}…`" send-label="Reply" cancellable autofocus @send="node.send" @cancel="node.cancel"/>
</div>
</div>
<template v-if="!collapsed"><template v-for="r in node.replies" :key="r.key"><CommentNode :node="r"/></template></template>
</div>
</template>

<script>
export default {
  name: 'CommentNode', props: { node: { type: Object, required: true } }, data() { return { collapsed: false }; },
  methods: {
    // the same small pill for Like, Dislike and Reply: lit up in its own colour while it is on
    chip(on, color) { return 'display: inline-flex; align-items: center; gap: 5px; height: 28px; padding: 0 10px; border-radius: 14px; cursor: pointer; font-size: 12px; font-weight: 600; background: ' + (on ? color + '22' : 'transparent') + '; color: ' + (on ? color : 'var(--k-9DB0A6,#9DB0A6)') + '; border: 1px solid ' + (on ? color + '66' : 'transparent'); }
  }
};
</script>
