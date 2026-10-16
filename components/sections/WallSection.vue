<!-- Community wall -->
<template><template v-if="show.feed">
<div style="width: 100%; display: flex; flex-wrap: wrap; gap: 24px; align-items: flex-start">

<div style="flex: 999 1 460px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Community wall</h1>
<p style="margin: 4px 0 0; color: #9DB0A6">Everyone's space to post freely: photos, news, shout-outs, polls.</p></div>

<div style="display: flex; gap: 10px; overflow-x: auto; padding-bottom: 4px">
<button style="flex: none; width: 112px; height: 180px; border-radius: 14px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; gap: 8px; padding: 14px 8px; box-sizing: border-box">
<span style="width: 40px; height: 40px; border-radius: 50%; background: #B8F36B; color: #0F1714; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg></span>
<span style="font-size: 13px; font-weight: 600">Add a moment</span>
</button>
<template v-for="(m, m__i) in moments" :key="m__i">
<button :style="`flex: none; width: 112px; height: 180px; border-radius: 14px; border: 0; background: ${m.bg}; color: #fff; cursor: pointer; display: flex; flex-direction: column; justify-content: space-between; padding: 10px; box-sizing: border-box; text-align: left`">
<span :style="`width: 34px; height: 34px; border-radius: 50%; border: 3px solid #B8F36B; background: ${m.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ m.initials }}</span>
<span style="font-size: 13px; font-weight: 600; line-height: 1.3">{{ m.caption }}</span>
</button>
</template>
</div>

<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; gap: 12px; align-items: flex-start">
<div aria-hidden="true" style="width: 42px; height: 42px; flex: none; border-radius: 50%; background: #C2185B; color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600">S</div>
<label style="flex: 1"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Write a post</span><textarea rows="2" :value="draft" @input="onDraft" placeholder="What's on your mind, Sam?" style="width: 100%; box-sizing: border-box; padding: 12px 14px; border-radius: 14px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 15px; resize: vertical"></textarea></label>
</div>
<template v-if="hasWFiles">
<div style="display: flex; flex-wrap: wrap; gap: 8px; padding-left: 54px">
<template v-for="(w, w__i) in wfiles" :key="w__i">
<div style="position: relative; display: flex; align-items: center; gap: 8px; padding: 6px; padding-right: 36px; border-radius: 12px; background: #121B17; border: 1px solid #2A3832; max-width: 240px">
<template v-if="w.isImg"><img :src="w.url" :alt="w.name" style="width: 52px; height: 52px; border-radius: 8px; object-fit: cover"></template>
<template v-if="w.notImg"><span style="width: 40px; height: 40px; border-radius: 8px; background: #22302A; color: #B8F36B; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M14 3v4h4"></path></svg></span></template>
<div style="min-width: 0; display: flex; flex-direction: column"><span style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ w.name }}</span><span style="font-size: 11px; color: #9DB0A6">{{ w.meta }}</span></div>
<button :aria-label="`Remove ${w.name}`" @click="w.remove" style="position: absolute; top: 4px; right: 4px; width: 28px; height: 28px; border-radius: 50%; border: 0; background: #22302A; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
</template>
</div>
</template>
<template v-if="emojiWallOpen">
<div role="menu" aria-label="Emoji" style="padding: 8px; border-radius: 14px; background: #1D2A25; border: 1px solid #34453D; display: grid; grid-template-columns: repeat(auto-fill, minmax(42px, 1fr)); gap: 2px">
<template v-for="(e, e__i) in emojisWall" :key="e__i"><button :aria-label="e.label" @click="e.add" style="height: 42px; border: 0; border-radius: 8px; background: transparent; font-size: 22px; cursor: pointer">{{ e.ch }}</button></template>
</div>
</template>
<div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center; border-top: 1px solid #24322C; padding-top: 10px">
<label style="position: relative; height: 38px; padding: 0 12px; border-radius: 10px; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><input type="file" accept="image/*,video/*" multiple="" aria-label="Add photo or video" @change="onWallFiles" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#6EE7C8" stroke-width="1.8"><rect x="3" y="5" width="18" height="14" rx="2"></rect><circle cx="9" cy="10" r="2"></circle><path d="M21 16l-5-5-8 8"></path></svg>Photo / video</label>
<label style="position: relative; height: 38px; padding: 0 12px; border-radius: 10px; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><input type="file" multiple="" aria-label="Attach a file" @change="onWallFiles" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#C9B6FF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg>File</label>
<button @click="toggleEmojiWall" style="height: 38px; padding: 0 12px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#FFD66B" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="9"></circle><path d="M8.5 14.5c1 1.2 2.2 1.8 3.5 1.8s2.5-.6 3.5-1.8M9 9.5h.01M15 9.5h.01"></path></svg>Emoji</button>
<button style="height: 38px; padding: 0 12px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#FFB86B" stroke-width="1.8" stroke-linecap="round"><path d="M5 20V10M12 20V4M19 20v-7"></path></svg>Poll</button>
<button style="height: 38px; padding: 0 12px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#7CC4FF" stroke-width="1.8"><rect x="3" y="5" width="18" height="16" rx="2"></rect><path d="M3 10h18M8 3v4M16 3v4"></path></svg>Event</button>
<span style="margin-left: auto; font-size: 12px; color: #9DB0A6">Visible to: Everyone</span>
<template v-if="canPostWall"><button @click="onPost" :style="postBtnStyle">Post</button></template><template v-if="cannotPostWall"><span style="font-size: 12px; color: #FFB86B">Your role can't post here</span></template>
</div>
</div>

<template v-for="(p, p__i) in posts" :key="p__i">
<article style="border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; overflow: hidden">
<div style="display: flex; gap: 12px; align-items: center; padding: 16px 16px 0">
<div aria-hidden="true" :style="`width: 42px; height: 42px; border-radius: 50%; background: ${p.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700`">{{ p.initials }}</div>
<div style="flex: 1; display: flex; flex-direction: column"><strong style="font-size: 15px">{{ p.author }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ p.meta }}</span></div>
<button aria-label="Post options" style="width: 36px; height: 36px; border-radius: 50%; border: 0; background: transparent; color: #9DB0A6; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><circle cx="5" cy="12" r="1.8"></circle><circle cx="12" cy="12" r="1.8"></circle><circle cx="19" cy="12" r="1.8"></circle></svg></button>
</div>
<p style="margin: 0; padding: 12px 16px; font-size: 15px; line-height: 1.55; white-space: pre-line">{{ p.text }}</p>
<template v-if="p.hasPhoto">
<div role="img" :aria-label="p.photoLabel" :style="`height: 300px; background: ${p.photoBg}; display: flex; align-items: flex-end; padding: 14px 16px; box-sizing: border-box`"><span style="font-size: 12px; color: rgba(255,255,255,0.85)">[PHOTO] {{ p.photoLabel }}</span></div>
</template>
<template v-if="p.hasPoster">
<div role="img" :aria-label="`Award poster: ${p.posterAward} for ${p.posterNames}`" style="margin: 0 16px 10px; padding: 26px 20px; border-radius: 14px; background: #24220F; border: 1px solid #4A4426; display: flex; flex-direction: column; align-items: center; gap: 10px; text-align: center">
<svg aria-hidden="true" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#FFD66B" stroke-width="1.6" stroke-linejoin="round"><path d="M8 4h8v5a4 4 0 0 1-8 0z"></path><path d="M8 6H5a3 3 0 0 0 3 4M16 6h3a3 3 0 0 1-3 4M12 13v4M8 20h8M10 17h4"></path></svg>
<span style="font-size: 12px; letter-spacing: 0.12em; text-transform: uppercase; color: #FFD66B; font-weight: 700">{{ p.posterAward }}</span>
<strong style="font-family: 'Space Grotesk', sans-serif; font-size: 22px; color: #F3EBC8; line-height: 1.3">{{ p.posterNames }}</strong>
<span style="font-size: 12px; color: #E3D9AE">[YOUR ORGANIZATION] · {{ p.posterWhen }}</span>
</div>
</template>
<template v-if="p.hasTags"><div style="padding: 0 16px 10px; display: flex; flex-wrap: wrap; gap: 6px; align-items: center"><span style="font-size: 12px; color: #9DB0A6">Tagged:</span><template v-for="(tg, tg__i) in p.tagList" :key="tg__i"><button @click="tg.open" style="height: 28px; padding: 0 10px; border-radius: 14px; border: 1px solid #4A4426; background: #24220F; color: #FFD66B; cursor: pointer; font-size: 12px; font-weight: 600">@{{ tg.name }}</button></template></div></template>
<template v-if="p.hasImages">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 2px">
<template v-for="(m, m__i) in p.images" :key="m__i"><img :src="m.url" :alt="m.name" style="width: 100%; height: 300px; object-fit: cover; display: block"></template>
</div>
</template>
<template v-if="p.hasDocs">
<div style="display: flex; flex-wrap: wrap; gap: 8px; padding: 0 16px 8px">
<template v-for="(m, m__i) in p.docs" :key="m__i"><div style="display: flex; gap: 10px; align-items: center; padding: 8px 12px; border-radius: 10px; background: #121B17; border: 1px solid #2A3832"><span style="width: 34px; height: 34px; border-radius: 8px; background: #22302A; color: #B8F36B; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M14 3v4h4"></path></svg></span><div style="display: flex; flex-direction: column"><strong style="font-size: 13px">{{ m.name }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ m.meta }} · Download</span></div></div></template>
</div>
</template>
<template v-if="p.hasPoll">
<div style="display: flex; flex-direction: column; gap: 8px; padding: 0 16px 12px">
<template v-for="(o, o__i) in p.poll" :key="o__i">
<button @click="o.vote" :style="o.style"><span>{{ o.label }}</span><span style="margin-left: auto; font-weight: 600">{{ o.pct }}</span></button>
</template>
<span style="font-size: 12px; color: #9DB0A6">{{ p.pollNote }}</span>
</div>
</template>
<div style="display: flex; justify-content: space-between; padding: 10px 16px; font-size: 13px; color: #9DB0A6"><span>{{ p.likes }} likes</span><span>{{ p.commentCount }} comments · {{ p.shares }} shares</span></div>
<div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 4px; margin: 0 12px; padding: 4px 0; border-top: 1px solid #24322C; border-bottom: 1px solid #24322C">
<button @click="p.like" :style="p.likeStyle"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" :fill="p.heartFill" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z"></path></svg>Like</button>
<button style="height: 40px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M4 5h16v11H8l-4 4z"></path></svg>Comment</button>
<button @click="p.toggleShare" :aria-expanded="p.shareOpen" style="height: 40px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M14 5l7 7-7 7M21 12H9a6 6 0 0 0-6 6"></path></svg>Share</button>
</div>
<template v-if="p.shareOpen">
<div style="margin: 10px 12px 0; padding: 12px; border-radius: 12px; background: #1A2520; border: 1px solid #2A3832; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 13px; color: #C9D6CF">Share this post to a forum</span>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-for="(t, t__i) in p.shareTargets" :key="t__i"><button @click="t.pick" style="height: 36px; padding: 0 12px 0 6px; border-radius: 18px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; gap: 8px; font-size: 13px"><span :style="`width: 24px; height: 24px; border-radius: 7px; background: ${t.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 10px`">{{ t.initials }}</span>{{ t.name }}</button></template>
<button @click="p.copyLink" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Copy link</button>
</div>
<span style="font-size: 13px; color: #C9D6CF; margin-top: 4px">Share outside Hub</span>
<div style="display: flex; gap: 10px; padding: 10px; border-radius: 12px; background: #121B17; border: 1px solid #2A3832">
<template v-if="p.hasPreviewImg"><img :src="p.previewImg" alt="" style="width: 84px; height: 84px; flex: none; object-fit: cover; border-radius: 8px"></template>
<template v-if="p.hasPreviewBg"><span aria-hidden="true" :style="`width: 84px; height: 84px; flex: none; border-radius: 8px; background: ${p.photoBg}`"></span></template>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 13px">{{ p.author }} on Hub</strong><span style="font-size: 13px; line-height: 1.4; color: #C9D6CF">{{ p.previewText }}</span><span style="font-size: 11px; color: #9DB0A6">hub.[yourdomain]/p/{{ p.id }}</span></div>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-if="p.shareOutBlocked"><span style="font-size: 12px; color: #FFB86B">Your role isn't allowed to share outside Hub.</span></template>
<template v-for="(so, so__i) in p.shareOut" :key="so__i"><button @click="so.run" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid #2A3832; background: #E8EFEA; color: #0F1714; cursor: pointer; font-size: 13px; font-weight: 600">{{ so.label }}</button></template>
</div>
<span style="font-size: 11px; color: #9DB0A6">The image and text above go with the link. Admins can turn off outside sharing for private forums.</span>
</div>
</template>
<template v-if="p.hasShareNote"><span style="margin: 8px 16px 0; font-size: 12px; color: #B8F36B">{{ p.shareNote }}</span></template>
<div style="display: flex; flex-direction: column; gap: 10px; padding: 12px 16px 16px">
<template v-for="(c, c__i) in p.commentList" :key="c__i">
<div style="display: flex; gap: 8px; align-items: flex-start">
<div aria-hidden="true" :style="`width: 32px; height: 32px; flex: none; border-radius: 50%; background: ${c.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ c.initials }}</div>
<div style="padding: 8px 12px; border-radius: 16px; background: #1F2C26"><strong style="font-size: 13px">{{ c.who }}</strong><div style="font-size: 14px; line-height: 1.4">{{ c.text }}</div></div>
</div>
</template>
<div style="display: flex; gap: 8px; align-items: center">
<div aria-hidden="true" style="width: 32px; height: 32px; flex: none; border-radius: 50%; background: #C2185B; color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 12px">S</div>
<label style="flex: 1"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Write a comment</span><input type="text" @keydown="p.onComment" placeholder="Write a comment… (Enter to send)" style="width: 100%; box-sizing: border-box; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
</div>
</div>
</article>
</template>
</div>

<aside aria-label="Wall highlights" class="hub-static" style="flex: 1 1 250px; min-width: 0; position: sticky; top: 0; align-self: flex-start; max-height: calc(100vh - 100px); overflow-y: auto; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<strong style="font-size: 15px">Birthdays today</strong>
<span style="font-size: 14px; line-height: 1.5; color: #C9D6CF">Faith Achieng and 2 others. Send them a wish on the wall.</span>
<button style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Write a wish</button>
</div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 15px">Online now</strong>
<template v-for="(m, m__i) in people" :key="m__i">
<div style="display: flex; gap: 10px; align-items: center">
<span :style="`position: relative; width: 34px; height: 34px; flex: none; border-radius: 50%; background: ${m.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ m.initials }}<span style="position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: #B8F36B; border: 2px solid #16201C"></span></span>
<span style="font-size: 14px">{{ m.name }}</span>
</div>
</template>
</div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px">
<strong style="font-size: 15px">Trending</strong>
<span style="font-size: 14px; color: #C9D6CF">#CharityRun</span>
<span style="font-size: 14px; color: #C9D6CF">#NewJoiners</span>
<span style="font-size: 14px; color: #C9D6CF">#TournamentNight</span>
</div>
</aside>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'WallSection',
  ...hubBindings(["canPostWall","cannotPostWall","draft","emojiWallOpen","emojisWall","hasWFiles","moments","onDraft","onPost","onWallFiles","people","postBtnStyle","posts","show","toggleEmojiWall","wfiles"])
};
</script>
