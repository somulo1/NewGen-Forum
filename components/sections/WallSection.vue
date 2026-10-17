<!-- Community wall -->
<template><template v-if="show.feed">
<div style="width: 100%; display: flex; flex-wrap: wrap; gap: 24px; align-items: flex-start">

<div style="flex: 999 1 460px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Community wall</h1>
<p style="margin: 4px 0 0; color: #9DB0A6">Everyone's space to post freely: photos, news, shout-outs, polls.</p></div>

<div style="display: flex; gap: 10px; overflow-x: auto; padding-bottom: 4px">
<button @click="momentForm.toggle" :aria-expanded="momentForm.open" style="flex: none; width: 112px; height: 180px; border-radius: 14px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; gap: 8px; padding: 14px 8px; box-sizing: border-box">
<span style="width: 40px; height: 40px; border-radius: 50%; background: #B8F36B; color: #0F1714; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg></span>
<span style="font-size: 13px; font-weight: 600">Add a moment</span>
</button>
<template v-for="m in moments" :key="m.key">
<button @click="m.open" :aria-label="`View moment: ${m.caption}`" :style="`position: relative; flex: none; width: 112px; height: 180px; padding: 0; border-radius: 14px; border: 0; overflow: hidden; background: ${m.bg}; color: #fff; cursor: pointer; text-align: left; box-sizing: border-box`">
<img v-if="m.mediaIsImg" :src="m.mediaUrl" alt="" style="position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover">
<video v-else-if="m.mediaIsVideo" :src="m.mediaUrl" muted playsinline preload="metadata" style="position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover"></video>
<span aria-hidden="true" style="position: absolute; inset: 0; background: linear-gradient(to bottom, rgba(0,0,0,.18), rgba(0,0,0,0) 35%, rgba(0,0,0,.7))"></span>
<span style="position: relative; display: flex; flex-direction: column; justify-content: space-between; height: 100%; padding: 10px; box-sizing: border-box; min-width: 0">
<span style="display: flex; align-items: center; justify-content: space-between; gap: 4px"><span :style="`width: 34px; height: 34px; flex: none; border-radius: 50%; border: 3px solid #B8F36B; background: ${m.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ m.initials }}</span><span v-if="m.mediaIsVideo" aria-label="Video" style="font-size: 11px; padding: 2px 6px; border-radius: 8px; background: rgba(0,0,0,.55)">▶</span></span>
<span style="display: flex; flex-direction: column; gap: 4px; min-width: 0">
<span style="font-size: 13px; font-weight: 600; line-height: 1.3; overflow-wrap: anywhere; word-break: break-word; display: -webkit-box; -webkit-line-clamp: 4; -webkit-box-orient: vertical; overflow: hidden">{{ m.caption }}</span>
<span v-if="m.likeCount + m.loveCount" style="font-size: 11px; opacity: .9; display: flex; align-items: center; gap: 4px"><svg aria-hidden="true" width="11" height="11" viewBox="0 0 24 24" fill="#FF6B81" stroke="#FF6B81" stroke-width="2" stroke-linejoin="round"><path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z"></path></svg>{{ m.likeCount + m.loveCount }}</span>
</span>
</span>
</button>
</template>
</div>
<template v-if="momentForm.open">
<form @submit.prevent="submitMoment" aria-label="Add a moment" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; padding: 12px; border-radius: 14px; background: #16201C; border: 1px solid #24322C">
<input ref="momentCaptionInput" v-model="momentCaption" type="text" maxlength="120" placeholder="What's happening? (optional with a photo or video)" aria-label="Moment caption" style="flex: 1 1 220px; height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<EmojiButton :size="38" @pick="insertMomentEmoji"/>
<span style="display: flex; gap: 6px"><template v-for="c in momentForm.colors" :key="c"><button type="button" :aria-label="`Background ${c}`" @click="momentBg = c" :style="`width: 26px; height: 26px; border-radius: 50%; border: 2px solid ${momentBg === c ? '#E8EFEA' : 'transparent'}; background: ${c}; cursor: pointer`"></button></template></span>
<label style="position: relative; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; display: flex; align-items: center; gap: 6px; cursor: pointer; font-size: 13px; max-width: 200px; overflow: hidden"><input type="file" accept="image/*,video/*" aria-label="Add a photo or video to your moment" @change="pickMomentFile" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ momentFile ? momentFile.name : 'Photo / video' }}</span></label>
<button type="submit" style="height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Share moment</button>
<button type="button" @click="momentForm.toggle" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button>
</form>
</template>
<MomentViewer />

<div v-if="!momentForm.open" style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; gap: 12px; align-items: flex-start">
<div aria-hidden="true" style="width: 42px; height: 42px; flex: none; border-radius: 50%; background: #C2185B; color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600">S</div>
<RichTextInput style="flex: 1; min-width: 0" :model-value="draft" @update:model-value="onDraftValue" placeholder="What's on your mind, Sam?" aria-label="Write a post" :rows="3" :input-style="`color: ${postBg ? '#FFFFFF' : '#E8EFEA'}; font-size: 15px`"/>
</div>
<div role="group" aria-label="Post background color" style="display: flex; flex-wrap: wrap; align-items: center; gap: 7px; padding-left: 54px">
<span style="font-size: 12px; color: #9DB0A6">Post color</span>
<button type="button" aria-label="Use default post background" :aria-pressed="!postBg" @click="postBg = ''" :style="`width: 24px; height: 24px; border-radius: 50%; border: 2px solid ${!postBg ? '#E8EFEA' : '#526259'}; background: #16201C; cursor: pointer`"></button>
<template v-for="c in momentForm.colors" :key="c"><button type="button" :aria-label="`Post background ${c}`" :aria-pressed="postBg === c" @click="postBg = c" :style="`width: 24px; height: 24px; border-radius: 50%; border: 2px solid ${postBg === c ? '#E8EFEA' : 'transparent'}; background: ${c}; cursor: pointer`"></button></template>
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
<template v-if="wpollView.on">
<div style="display: flex; flex-direction: column; gap: 8px; padding: 12px; border-radius: 12px; background: #121B17; border: 1px solid #2A3832">
<span style="font-size: 12px; color: #9DB0A6">Poll: your post text is the question. Add 2 to 5 options.</span>
<template v-for="o in wpollView.opts" :key="o.n"><div style="display: flex; gap: 6px"><input type="text" :value="o.value" @input="o.onInput" :placeholder="`Option ${o.n}`" :aria-label="`Poll option ${o.n}`" style="flex: 1; min-width: 0; height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><button v-if="o.canRemove" type="button" :aria-label="`Remove option ${o.n}`" @click="o.remove" style="width: 38px; height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div></template>
<button v-if="wpollView.canAdd" type="button" @click="wpollView.add" style="align-self: flex-start; height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">+ Add option</button>
</div>
</template>
<div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center; border-top: 1px solid #24322C; padding-top: 10px">
<label style="position: relative; height: 38px; padding: 0 12px; border-radius: 10px; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><input type="file" accept="image/*,video/*" multiple="" aria-label="Add photo or video" @change="onWallFiles" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#6EE7C8" stroke-width="1.8"><rect x="3" y="5" width="18" height="14" rx="2"></rect><circle cx="9" cy="10" r="2"></circle><path d="M21 16l-5-5-8 8"></path></svg>Photo / video</label>
<label style="position: relative; height: 38px; padding: 0 12px; border-radius: 10px; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><input type="file" multiple="" aria-label="Attach a file" @change="onWallFiles" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#C9B6FF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg>File</label>
<EmojiButton :size="38" @pick="pickEmojiWall"/>
<button @click="wpollView.toggle" :aria-pressed="wpollView.on" style="height: 38px; padding: 0 12px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#FFB86B" stroke-width="1.8" stroke-linecap="round"><path d="M5 20V10M12 20V4M19 20v-7"></path></svg>Poll</button>
<button @click="openEventFromWall" style="height: 38px; padding: 0 12px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#7CC4FF" stroke-width="1.8"><rect x="3" y="5" width="18" height="16" rx="2"></rect><path d="M3 10h18M8 3v4M16 3v4"></path></svg>Event</button>
<template v-if="canPostWall"><button @click="submitWallPost" :style="postBtnStyle">Post</button></template><template v-if="cannotPostWall"><span style="font-size: 12px; color: #FFB86B">Your role can't post here</span></template>
</div>
</div>

<template v-if="wallFilter.active"><div style="display: flex; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 12px; background: #1B2A1E; font-size: 14px">Showing posts about <strong>#{{ wallFilter.tag }}</strong><button @click="wallFilter.clear" style="margin-left: auto; height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Clear</button></div></template>
<template v-if="wallFilter.active && !posts.length"><p style="margin: 0; color: #9DB0A6">No posts match this hashtag.</p></template>
<template v-for="(p, p__i) in posts" :key="p__i">
<article style="border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; overflow: hidden">
<div style="display: flex; gap: 12px; align-items: center; padding: 16px 16px 0">
<div aria-hidden="true" :style="`width: 42px; height: 42px; border-radius: 50%; background: ${p.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700`">{{ p.initials }}</div>
<div style="flex: 1; display: flex; flex-direction: column"><strong style="font-size: 15px">{{ p.author }}<template v-if="p.saved"> <span style="font-size: 11px; color: #B8F36B; font-weight: 500">· Saved</span></template></strong><span style="font-size: 12px; color: #9DB0A6">{{ p.meta }}</span></div>
<div style="position: relative"><button aria-label="Post options" aria-haspopup="menu" :aria-expanded="p.menuOpen" @click="p.toggleMenu" style="width: 36px; height: 36px; border-radius: 50%; border: 0; background: transparent; color: #9DB0A6; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><circle cx="5" cy="12" r="1.8"></circle><circle cx="12" cy="12" r="1.8"></circle><circle cx="19" cy="12" r="1.8"></circle></svg></button>
<template v-if="p.menuOpen"><div role="menu" aria-label="Post options" style="position: absolute; right: 0; top: 40px; z-index: 5; min-width: 170px; padding: 6px; border-radius: 12px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 14px 30px rgba(0,0,0,0.45); display: flex; flex-direction: column"><template v-for="(mi, mi__i) in p.menuItems" :key="mi__i"><button role="menuitem" @click="mi.run" :style="`height: 38px; padding: 0 12px; border: 0; border-radius: 8px; background: transparent; text-align: left; cursor: pointer; font-size: 14px; color: ${mi.danger ? '#FF9B8A' : '#E8EFEA'}`">{{ mi.label }}</button></template></div></template></div>
</div>
<p v-html="p.htmlText" :style="`display: ${p.postBg ? 'flex' : 'block'}; align-items: center; justify-content: center; box-sizing: border-box; min-height: ${p.postBg ? '220px' : '0'}; margin: 0${p.postBg ? ' 16px 10px' : ''}; padding: ${p.postBg ? '28px 24px' : '12px 16px'}; border-radius: ${p.postBg ? '12px' : '0'}; background: ${p.postBg || 'transparent'}; color: ${p.postBg ? '#FFFFFF' : '#E8EFEA'}; text-align: ${p.postBg ? 'center' : 'left'}; font-size: ${p.postBg ? '22px' : '15px'}; font-weight: ${p.postBg ? '600' : '400'}; line-height: 1.45; white-space: pre-line; overflow-wrap: anywhere`"></p>
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
<template v-if="p.hasMedia"><div style="padding: 0 16px 12px"><MediaCollage :files="p.media"/></div></template>
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
<button @click="p.toggleComments" :aria-expanded="p.commentsOpen" style="height: 40px; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M4 5h16v11H8l-4 4z"></path></svg>Comment</button>
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
<template v-if="p.commentsOpen">
<div role="region" aria-label="Comments" style="border-top: 1px solid #24322C; display: flex; flex-direction: column; min-width: 0">
<div style="max-height: 340px; overflow-y: auto; padding: 4px 16px 8px">
<template v-if="!p.hasComments"><p style="margin: 12px 0; font-size: 13px; color: #9DB0A6">No comments yet. Be the first.</p></template>
<template v-for="n in p.commentTree" :key="n.key"><CommentNode :node="n"/></template>
</div>
<div v-if="!p.replyingAny" style="display: flex; gap: 8px; align-items: center; padding: 8px 16px 14px; border-top: 1px solid #24322C">
<div aria-hidden="true" style="width: 32px; height: 32px; flex: none; border-radius: 50%; background: #C2185B; color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 12px">S</div>
<CommentBox style="flex: 1" placeholder="Write a comment…" @send="p.addComment"/>
</div>
</div>
</template>
</article>
</template>
</div>

<aside aria-label="Wall highlights" class="hub-static" style="flex: 1 1 250px; min-width: 0; position: sticky; top: 0; align-self: flex-start; overflow: visible; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<strong style="font-size: 15px">Birthdays today</strong>
<span style="font-size: 14px; line-height: 1.5; color: #C9D6CF">Faith Achieng and 2 others. Send them a wish on the wall.</span>
<button @click="wish" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Write a wish</button>
</div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 15px">Online now</strong>
<template v-for="(m, m__i) in onlineList" :key="m__i">
<button @click="m.open" :aria-label="`View ${m.name}`" style="display: flex; gap: 10px; align-items: center; padding: 0; border: 0; background: transparent; color: inherit; cursor: pointer; text-align: left; font-family: inherit">
<span :style="`position: relative; width: 34px; height: 34px; flex: none; border-radius: 50%; background: ${m.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ m.initials }}<span style="position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: #B8F36B; border: 2px solid #16201C"></span></span>
<span style="font-size: 14px">{{ m.name }}</span>
</button>
</template>
</div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px">
<strong style="font-size: 15px">Trending</strong>
<template v-for="tg in trending" :key="tg.tag"><button @click="tg.pick" :aria-pressed="tg.on" :style="`text-align: left; padding: 6px 8px; margin: 0 -8px; border: 0; border-radius: 8px; background: ${tg.on ? '#22302A' : 'transparent'}; color: ${tg.on ? '#B8F36B' : '#C9D6CF'}; cursor: pointer; font-size: 14px; font-family: inherit`">#{{ tg.tag }}</button></template>
</div>
</aside>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'WallSection',
  data() { return { momentCaption: '', momentBg: '#2B4A63', momentFile: null, postBg: '' }; },
  methods: {
    submitMoment() { if (this.addMoment(this.momentCaption, this.momentBg, this.momentFile)) { this.momentCaption = ''; this.momentFile = null; } },
    submitWallPost() { if (this.onPost(this.postBg)) this.postBg = ''; },
    insertMomentEmoji(emoji) {
      const input = this.$refs.momentCaptionInput;
      if (!input) return;
      const start = input.selectionStart ?? this.momentCaption.length;
      const end = input.selectionEnd ?? start;
      this.momentCaption = (this.momentCaption.slice(0, start) + emoji + this.momentCaption.slice(end)).slice(0, 120);
      this.$nextTick(() => {
        input.focus();
        const caret = Math.min(start + emoji.length, 120);
        input.setSelectionRange(caret, caret);
      });
    },
    pickMomentFile(e) { const f = e.target.files && e.target.files[0]; e.target.value = ''; if (f) this.momentFile = f; },
    wish() { this.wishBirthday(); this.$nextTick(() => this.$refs.composer && this.$refs.composer.focus()); }
  },
  ...hubBindings(["pickEmojiWall","addMoment","momentForm","momentView","wpollView","openEventFromWall","wallAudience","wallAudiences","onAudience","wishBirthday","onlineList","trending","wallFilter","canPostWall","cannotPostWall","draft","hasWFiles","moments","onDraftValue","onPost","onWallFiles","people","postBtnStyle","posts","show","wfiles"])
};
</script>
