<!-- Projects list and project view -->
<template><template v-if="showProjectList">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Projects</h1><p style="margin: 4px 0 0; color: #9DB0A6">Every project's goal, progress, milestones and sprints in one place.</p></div>
<template v-if="canCreateProject"><button @click="openProjectForm" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">New project</button></template>
</div>
<template v-if="projectForm.open">
<form novalidate @submit.prevent="projectForm.submit" aria-label="Create project" style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #B8F36B; display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 16px">Create a project</strong>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Project name<input type="text" :value="projectForm.name" @input="projectForm.onName" placeholder="e.g. Customer self-service portal" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Goal<textarea rows="3" :value="projectForm.goal" @input="projectForm.onGoal" placeholder="What outcome should this project deliver?" style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; line-height: 1.45; resize: vertical"></textarea></label>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(190px, 1fr)); gap: 12px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Project owner<select :value="projectForm.owner" @change="projectForm.onOwner" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><template v-for="(o, o__i) in projectForm.owners" :key="o__i"><option :value="o.v">{{ o.l }}</option></template></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Estimated budget<input type="number" min="0" :value="projectForm.budget" @input="projectForm.onBudget" placeholder="Optional" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
</div>
<div style="display: flex; flex-direction: column; gap: 8px">
<label for="project-team-search" style="font-size: 12px; color: #9DB0A6">Project team</label>
<input id="project-team-search" type="search" :value="projectForm.teamQuery" @input="projectForm.onTeamQuery" placeholder="Search people by name…" style="height: 40px; box-sizing: border-box; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<div v-if="projectForm.team.length" style="display: flex; flex-wrap: wrap; gap: 6px">
<span v-for="member in projectForm.selectedTeam" :key="member.name" style="display: flex; align-items: center; gap: 6px; min-height: 30px; padding: 2px 6px; border-radius: 16px; background: #22302A; color: #E8EFEA; font-size: 12px">
<UserAvatar :name="member.name" :initials="member.initials" :color="member.color" :size="22" :font-size="8"/>
{{ member.name }}
<button v-if="member.removable" type="button" :aria-label="`Remove ${member.name} from project team`" @click="projectForm.removeTeamMember(member.name)" style="padding: 0 4px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; font-size: 16px">×</button>
</span>
</div>
<div style="max-height: 180px; overflow-y: auto; border: 1px solid #2A3832; border-radius: 10px; background: #121B17">
<label v-for="member in projectForm.teamResults" :key="member.name" style="display: flex; align-items: center; gap: 9px; min-height: 40px; padding: 0 10px; border-bottom: 1px solid #1D2A25; cursor: pointer; font-size: 13px">
<input type="checkbox" :checked="member.selected" @change="member.toggle" style="accent-color: #B8F36B">
<UserAvatar :name="member.name" :initials="member.initials" :color="member.color" :size="26" :font-size="9"/>
{{ member.name }}
</label>
<span v-if="!projectForm.teamResults.length" style="display: block; padding: 12px; color: #9DB0A6; font-size: 13px">No people match that search.</span>
</div>
<span style="font-size: 11px; color: #9DB0A6">Select multiple people. The project owner is always included.</span>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(190px, 1fr)); gap: 12px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Start date<input type="date" :value="projectForm.start" @input="projectForm.onStart" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Due date<input type="date" :value="projectForm.due" :min="projectForm.start" @input="projectForm.onDue" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
</div>
<span v-if="projectForm.err" role="alert" style="font-size: 13px; color: #FF9B8A">{{ projectForm.err }}</span>
<div style="display: flex; flex-wrap: wrap; gap: 8px"><button type="submit" style="height: 40px; padding: 0 20px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Submit project</button><button type="button" @click="projectForm.cancel" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button></div>
</form>
</template>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px">
<template v-for="(ps, ps__i) in projStats" :key="ps__i">
<div style="padding: 14px 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 28px; font-weight: 700; color: ${ps.color}`">{{ ps.n }}</span><span style="font-size: 13px; color: #C9D6CF">{{ ps.label }}</span></div>
</template>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 12px; align-items: center; padding: 12px 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C">
<span style="width: 34px; height: 34px; border-radius: 9px; background: #22302A; color: #E8EFEA; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 16px">#</span>
<div style="flex: 1 1 260px; display: flex; flex-direction: column"><strong style="font-size: 14px">Slack · connected</strong><span style="font-size: 12px; color: #9DB0A6">Each project posts task moves, overdue reminders and milestones to its own Slack channel.</span></div>
<span style="padding: 4px 10px; border-radius: 12px; background: #26361A; color: #B8F36B; font-size: 12px; font-weight: 600">Active</span>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 14px">
<template v-for="(pr, pr__i) in projects" :key="pr__i">
<article style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; gap: 12px; align-items: center">
<span :style="`width: 42px; height: 42px; flex: none; border-radius: 12px; background: ${pr.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 14px`">{{ pr.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 16px">{{ pr.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ pr.code }} · Led by {{ pr.owner }}</span></div>
<span :style="pr.statusStyle">{{ pr.status }}</span>
</div>
<p style="margin: 0; font-size: 13px; line-height: 1.5; color: #C9D6CF">{{ pr.goal }}</p>
<div style="display: flex; flex-direction: column; gap: 6px">
<div style="display: flex; justify-content: space-between; font-size: 12px"><span style="color: #9DB0A6">Progress</span><strong>{{ pr.progress }}%</strong></div>
<div style="height: 8px; border-radius: 4px; background: #22302A; overflow: hidden"><div :style="pr.barStyle"></div></div>
</div>
<div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px; font-size: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><span style="color: #9DB0A6">Next milestone</span><strong style="font-weight: 600">{{ pr.nextMilestone }}</strong></div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="color: #9DB0A6">Due</span><strong style="font-weight: 600">{{ pr.dueLabel }}</strong></div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="color: #9DB0A6">Sprint</span><strong style="font-weight: 600">{{ pr.sprintLabel }}</strong></div>
</div>
<div style="display: flex; align-items: center; gap: 10px">
<div style="display: flex">
<template v-for="(tm, tm__i) in pr.team" :key="tm__i"><UserAvatar :name="tm.name" :initials="tm.initials" :color="tm.color" :size="30" :font-size="10" border="2px solid #16201C" style="margin-right: -8px"/></template>
</div>
<span style="font-size: 12px; color: #9DB0A6; margin-left: 8px">{{ pr.teamSize }} people · {{ pr.sync }}</span>
<button @click="pr.open" style="margin-left: auto; height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">View</button>
</div>
</article>
</template>
</div>
</div>
</template>

<template v-if="showProjectView">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<button @click="closeProject" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; gap: 6px"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg>All projects</button>

<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: center; padding: 20px; border-radius: 20px; background: #16201C; border: 1px solid #24322C">
<span :style="`width: 60px; height: 60px; flex: none; border-radius: 16px; background: ${pj.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 20px`">{{ pj.initials }}</span>
<div style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ pj.name }}</h1><span :style="pj.statusStyle">{{ pj.status }}</span></div>
<span style="font-size: 13px; color: #9DB0A6">{{ pj.code }} · Led by {{ pj.owner }} · {{ pj.startLabel }} → {{ pj.dueLabel }}</span>
</div>
<template v-if="has.sessions">
<div style="display: flex; gap: 8px; flex-wrap: wrap; align-items: center">
<template v-if="pj.standup.isLive"><span style="font-size: 12px; font-weight: 600; color: #FF9B8A">● LIVE · {{ pj.standup.label }}</span><button @click="pj.standup.join" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Join standup</button></template>
<template v-if="pj.standup.isUpcoming"><span style="height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; display: flex; align-items: center; font-size: 13px; color: #C9D6CF">Standup · {{ pj.standup.label }}</span><button @click="pj.standup.cancel" style="height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button></template>
<template v-if="pj.standup.isNone"><button @click="pj.standup.openForm" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Schedule standup</button></template>
</div>
</template>
</div>

<template v-if="pj.standup.formOpen">
<form novalidate @submit.prevent="pj.standup.form.submit" aria-label="Schedule standup" style="padding: 16px; border-radius: 14px; background: #121B17; border: 1px solid #B8F36B; display: flex; flex-wrap: wrap; gap: 12px; align-items: flex-end">
<strong style="flex: 1 1 100%; font-size: 14px">Schedule a standup for the whole team</strong>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Date<input type="date" :value="pj.standup.form.date" :min="pj.standup.form.min" @input="pj.standup.form.onDate" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Start time<input type="time" :value="pj.standup.form.start" @input="pj.standup.form.onStart" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Repeats<select :value="pj.standup.form.repeat" @change="pj.standup.form.onRepeat" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><option value="None">Once</option><option value="Weekdays">Weekdays</option><option value="Weekly">Weekly</option></select></label>
<div style="display: flex; gap: 8px"><button type="submit" style="height: 40px; padding: 0 20px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Schedule</button><button type="button" @click="pj.standup.form.close" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button></div>
<template v-if="pj.standup.form.hasErr"><span role="alert" style="flex: 1 1 100%; font-size: 13px; color: #FF9B8A">{{ pj.standup.form.err }}</span></template>
<span style="flex: 1 1 100%; font-size: 12px; color: #9DB0A6">The standup lasts 15 minutes and is added to everyone's calendar. When the start time is reached, this page shows "Join standup".</span>
</form>
</template>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px">
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Overall progress</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.progress }}%</span><div style="height: 6px; border-radius: 3px; background: #22302A; overflow: hidden"><div :style="pj.barStyle"></div></div></div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Tasks done</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.doneCount }}<span style="font-size: 16px; color: #9DB0A6"> / {{ pj.taskCount }}</span></span></div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Milestones reached</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.msDone }}<span style="font-size: 16px; color: #9DB0A6"> / {{ pj.msCount }}</span></span></div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Time left</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.daysLeft }}<span style="font-size: 16px; color: #9DB0A6"> days</span></span></div>
</div>

<div style="padding: 16px 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px">
<div role="tablist" style="display: flex; padding: 3px; border-radius: 12px; background: #121B17; border: 1px solid #24322C"><button role="tab" @click="pj.goMs" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; ${pj.isMsTab ? 'background: #E8EFEA; color: #0F1714' : 'background: transparent; color: #C9D6CF'}`">Milestones ({{ pj.msCount }})</button><button role="tab" @click="pj.goSp" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; ${pj.isSpTab ? 'background: #E8EFEA; color: #0F1714' : 'background: transparent; color: #C9D6CF'}`">Sprints ({{ pj.spCount }})</button></div>
<span style="flex: 1 1 200px; font-size: 12px; color: #9DB0A6">Tasks sit in sprints, sprints belong to milestones, and progress rolls up automatically.</span>
<template v-if="pj.canMilestones"><button @click="pj.addSprint" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer; font-size: 13px">Add sprint</button><button @click="pj.addMilestone" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 13px">Add milestone</button></template>
</div>
<template v-if="pj.isMsTab"><DataTable :t="pj.msTable" /></template>
<template v-if="pj.isSpTab"><DataTable :t="pj.spTable" /></template>
</div>

<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
<div style="flex: 999 1 560px; min-width: 0; padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><strong style="flex: 1; font-size: 15px">Task board</strong>
<label style="display: flex; align-items: center; gap: 8px; font-size: 12px; color: #9DB0A6">Sprint<select :value="pj.sprintValue" @change="pj.setSprint" aria-label="Filter tasks by sprint" style="height: 38px; max-width: 260px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px"><template v-for="(sp, sp__i) in pj.sprintOpts" :key="sp__i"><option :value="sp.v">{{ sp.l }}</option></template></select></label>
<template v-if="pj.canNewTask"><button @click="pj.openTaskForm" style="height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 14px">New task</button></template>
</div>
<template v-if="pj.taskForm.open">
<form novalidate @submit.prevent="pj.taskForm.submit" aria-label="New task" style="padding: 16px; border-radius: 14px; background: #121B17; border: 1px solid #B8F36B; display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 14px">New task</strong>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Title<input type="text" :value="pj.taskForm.title" @input="pj.taskForm.onTitle" placeholder="What needs to be done?" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Description<textarea rows="3" :value="pj.taskForm.desc" @input="pj.taskForm.onDesc" placeholder="Add the details someone needs to do this task" style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px; line-height: 1.45; resize: vertical"></textarea></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Objective<input type="text" :value="pj.taskForm.objective" @input="pj.taskForm.onObjective" placeholder="What should this task achieve?" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Sprint<select :value="pj.taskForm.sprint" @change="pj.taskForm.onSprint" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><option value="">Choose a sprint…</option><template v-for="(so, so__i) in pj.taskForm.sprintOpts" :key="so__i"><option :value="so.v">{{ so.l }}</option></template></select></label>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Owner<select :value="pj.taskForm.owner" @change="pj.taskForm.onOwner" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><template v-for="(ow, ow__i) in pj.taskForm.ownerOpts" :key="ow__i"><option :value="ow.v">{{ ow.l }}</option></template></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Due date<input type="date" :value="pj.taskForm.due" :min="pj.taskForm.minDue" :max="pj.taskForm.maxDue" @input="pj.taskForm.onDue" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
</div>
<div style="padding: 10px 12px; border-radius: 10px; background: #16201C"><span style="font-size: 12px; color: #C9D6CF">{{ pj.taskForm.planLine }}</span></div>
<template v-if="pj.taskForm.noSprints"><div role="alert" style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center; font-size: 13px; color: #FFB86B">No open sprint exists yet.<button type="button" @click="pj.taskForm.addSprint" style="height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid #FFB86B; background: transparent; color: #FFB86B; cursor: pointer">Add a sprint</button></div></template>
<template v-if="pj.taskForm.hasErr"><span role="alert" style="font-size: 13px; color: #FF9B8A">{{ pj.taskForm.err }}</span></template>
<div style="display: flex; gap: 8px"><button type="submit" style="height: 40px; padding: 0 20px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Create task</button><button type="button" @click="pj.taskForm.cancel" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button></div>
</form>
</template>
<div style="display: grid; grid-template-columns: repeat(4, minmax(170px, 1fr)); gap: 10px; overflow-x: auto">
<template v-for="(col, col__i) in pj.columns" :key="col__i">
<div style="min-width: 170px; padding: 10px; border-radius: 14px; background: #121B17; display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; align-items: center; gap: 8px; padding: 2px 4px 4px"><span :style="`width: 8px; height: 8px; border-radius: 50%; background: ${col.color}`"></span><strong style="font-size: 13px; flex: 1">{{ col.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ col.count }}</span></div>
<div class="hub-task-column-scroll" tabindex="0" :aria-label="`${col.name} tasks`" :style="`display:flex;flex-direction:column;gap:8px;max-height:${col.count >= 5 ? '480px' : 'none'};overflow-y:${col.count >= 5 ? 'auto' : 'visible'};overscroll-behavior:contain`">
<template v-for="(tk2, tk2__i) in col.tasks" :key="tk2__i">
<div :style="tk2.cardStyle">
<button @click="tk2.open" style="padding: 0; border: 0; background: transparent; color: #E8EFEA; text-align: left; cursor: pointer; font-size: 13px; font-weight: 500; line-height: 1.35">{{ tk2.title }}</button>
<div style="display: flex; flex-wrap: wrap; gap: 4px"><span :style="tk2.dueStyle">{{ tk2.dueLabel }}</span><span :title="tk2.ctxTitle" style="padding: 2px 7px; border-radius: 6px; background: #1B2A33; color: #7CC4FF; font-size: 10px; font-weight: 600">{{ tk2.ctxLabel }}</span><template v-if="tk2.hasLinks"><span style="padding: 2px 7px; border-radius: 6px; background: #22302A; color: #C9D6CF; font-size: 10px; font-weight: 600">{{ tk2.linkLabel }}</span></template></div>
<div style="display: flex; align-items: center; gap: 4px">
<UserAvatar :name="tk2.primary" :initials="tk2.initials" :color="tk2.color" :size="24" :font-size="9" :aria-label="`Primary owner: ${tk2.primary}`"/>
<template v-if="tk2.hasSecondary"><UserAvatar :name="tk2.secondary" :initials="tk2.secInitials" :color="tk2.secColor" :size="18" :font-size="7" border="2px solid #1D2A25" :aria-label="`Secondary owner: ${tk2.secondary}`" style="margin-left: -8px"/></template>
<span style="flex: 1"></span>
<template v-if="tk2.canBack"><button :aria-label="tk2.backLabel" :title="tk2.backLabel" @click="tk2.back" style="width: 28px; height: 28px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg></button></template>
<template v-if="tk2.canMove"><button :aria-label="tk2.moveLabel" :title="tk2.moveLabel" @click="tk2.move" style="width: 28px; height: 28px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6"></path></svg></button></template>
</div>
</div>
</template>
</div>
</div>
</template>
</div>
</div>

<aside aria-label="Project overview" style="flex: 1 1 300px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Why this project</span>
<p style="margin: 0; font-size: 14px; line-height: 1.55">{{ pj.goal }}</p>
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; margin-top: 6px">Team · {{ pj.teamCount }}</span>
<div style="max-height: 200px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; padding-right: 4px">
<template v-for="(tm, tm__i) in pj.team" :key="tm__i"><div style="display: flex; gap: 10px; align-items: center"><UserAvatar :name="tm.name" :initials="tm.initials" :color="tm.color" :size="30" :font-size="11"/><span style="font-size: 14px; flex: 1">{{ tm.name }}</span><span style="font-size: 12px; color: #9DB0A6">{{ tm.load }} open</span></div></template>
</div>
</div>
<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Recent activity</span>
<div style="max-height: 230px; overflow-y: auto; display: flex; flex-direction: column; gap: 10px; padding-right: 4px">
<template v-for="(pa, pa__i) in pj.activity" :key="pa__i"><div style="display: flex; gap: 10px; align-items: flex-start"><span style="width: 6px; height: 6px; margin-top: 7px; flex: none; border-radius: 50%; background: #B8F36B"></span><span style="font-size: 13px; line-height: 1.45; color: #C9D6CF">{{ pa.text }} <span style="color: #9DB0A6">· {{ pa.time }}</span></span></div></template>
</div>
</div>
</aside>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ProjectsSection',
  ...hubBindings(["canCreateProject","closeProject","has","openProjectForm","pj","projectForm","projStats","projects","showProjectList","showProjectView"])
};
</script>
