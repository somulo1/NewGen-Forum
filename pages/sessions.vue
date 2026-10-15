<template>
  <div class="sessions-container">
    <div class="page-header">
      <div>
        <h1>Meetings & Sessions</h1>
        <p class="subtitle">Video calls and team meetings</p>
      </div>
      <button class="btn btn-primary">
        <Icon name="calendar" size="18" />
        Schedule Meeting
      </button>
    </div>

    <div class="tabs">
      <button :class="['tab-btn', { active: activeTab === 'active' }]" @click="activeTab = 'active'">
        <Icon name="video" size="16" />
        Active Meetings
      </button>
      <button :class="['tab-btn', { active: activeTab === 'upcoming' }]" @click="activeTab = 'upcoming'">
        <Icon name="calendar" size="16" />
        Upcoming ({{ upcomingMeetings.length }})
      </button>
      <button :class="['tab-btn', { active: activeTab === 'past' }]" @click="activeTab = 'past'">
        <Icon name="clock" size="16" />
        Past Meetings
      </button>
    </div>

    <div v-if="activeTab === 'active'" class="meetings-grid">
      <div v-for="meeting in activeMeetings" :key="meeting.id" class="meeting-card active-meeting">
        <div class="video-preview">
          <div class="video-placeholder">
            <Icon name="video" size="48" />
            <span>Live Meeting</span>
          </div>
          <div class="participants">
            <div v-for="participant in meeting.participants" :key="participant.id" class="participant-avatar">
              {{ participant.initials }}
            </div>
          </div>
        </div>
        <div class="meeting-info">
          <h3>{{ meeting.title }}</h3>
          <p>{{ meeting.participants.length }} participants • Started {{ meeting.startTime }}</p>
        </div>
        <div class="meeting-actions">
          <button class="action-btn">Join</button>
          <button class="action-btn secondary">Details</button>
        </div>
      </div>

      <div v-if="activeMeetings.length === 0" class="no-content">
        <Icon name="video" size="48" />
        <p>No active meetings right now</p>
      </div>
    </div>

    <div v-if="activeTab === 'upcoming'" class="meetings-list">
      <div v-for="meeting in upcomingMeetings" :key="meeting.id" class="meeting-item">
        <div class="meeting-time">
          <div class="time-box">
            <span class="time-hour">{{ meeting.time }}</span>
            <span class="time-date">{{ meeting.date }}</span>
          </div>
        </div>
        <div class="meeting-content">
          <h3>{{ meeting.title }}</h3>
          <p>{{ meeting.organizer }} • {{ meeting.participants }} participants</p>
          <div class="meeting-description">{{ meeting.description }}</div>
        </div>
        <div class="meeting-badges">
          <span v-if="meeting.recurring" class="badge">Recurring</span>
          <span v-if="meeting.recorded" class="badge">Recording enabled</span>
        </div>
        <button class="btn-join">Join Meeting</button>
      </div>
    </div>

    <div v-if="activeTab === 'past'" class="meetings-list">
      <div v-for="meeting in pastMeetings" :key="meeting.id" class="meeting-item past">
        <div class="meeting-time">
          <div class="time-box">
            <span class="time-hour">{{ meeting.time }}</span>
            <span class="time-date">{{ meeting.date }}</span>
          </div>
        </div>
        <div class="meeting-content">
          <h3>{{ meeting.title }}</h3>
          <p>{{ meeting.organizer }} • {{ meeting.duration }} minutes</p>
        </div>
        <div class="meeting-actions">
          <button v-if="meeting.recording" class="btn-recording">
            <Icon name="play" size="14" />
            View Recording
          </button>
          <button class="btn-details">Details</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const activeTab = ref('active')

const activeMeetings = [
  {
    id: 1,
    title: 'Team Standup',
    startTime: '10:05 AM',
    participants: [
      { id: 1, initials: 'JD' },
      { id: 2, initials: 'SM' },
      { id: 3, initials: 'MD' },
      { id: 4, initials: 'EW' },
    ],
  },
]

const upcomingMeetings = [
  {
    id: 1,
    title: 'Project Review Meeting',
    time: '2:00 PM',
    date: 'Today',
    organizer: 'Sarah Martin',
    participants: 8,
    description: 'Q4 project progress review and planning',
    recurring: true,
    recorded: true,
  },
  {
    id: 2,
    title: 'Client Call',
    time: '3:30 PM',
    date: 'Today',
    organizer: 'John Doe',
    participants: 4,
    description: 'Quarterly business review',
    recurring: false,
    recorded: true,
  },
  {
    id: 3,
    title: 'Team Building Session',
    time: '4:00 PM',
    date: 'Tomorrow',
    organizer: 'HR Team',
    participants: 25,
    description: 'Virtual team building activities',
    recurring: false,
    recorded: false,
  },
]

const pastMeetings = [
  {
    id: 1,
    title: 'Daily Standup',
    time: '10:00 AM',
    date: 'Dec 14',
    organizer: 'Marcus Davis',
    duration: 15,
    recording: true,
  },
  {
    id: 2,
    title: 'Design Review',
    time: '2:00 PM',
    date: 'Dec 13',
    organizer: 'Sarah Martin',
    duration: 45,
    recording: true,
  },
  {
    id: 3,
    title: 'All Hands Meeting',
    time: '1:00 PM',
    date: 'Dec 12',
    organizer: 'Executive Team',
    duration: 60,
    recording: false,
  },
]
</script>

<style scoped>
.sessions-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: var(--spacing-sm) 0 0 0;
}

.tabs {
  display: flex;
  gap: var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md) var(--spacing-lg);
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-weight: 500;
  font-size: 13px;
  border-bottom: 3px solid transparent;
  transition: all 0.2s ease;
  margin-bottom: -1px;
}

.tab-btn:hover {
  color: var(--color-text);
}

.tab-btn.active {
  color: var(--color-lime);
  border-bottom-color: var(--color-lime);
}

.meetings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-lg);
}

.meeting-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all 0.2s ease;
}

.meeting-card:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.video-preview {
  position: relative;
  background: var(--color-bg);
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  color: var(--color-text-secondary);
}

.video-placeholder :deep(svg) {
  opacity: 0.5;
}

.participants {
  position: absolute;
  bottom: var(--spacing-md);
  right: var(--spacing-md);
  display: flex;
  gap: -8px;
  flex-direction: row-reverse;
}

.participant-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 11px;
  border: 2px solid var(--color-bg);
  margin-left: -8px;
}

.meeting-info {
  padding: var(--spacing-lg);
}

.meeting-info h3 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: 15px;
  font-weight: 600;
}

.meeting-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.meeting-actions {
  padding: 0 var(--spacing-lg) var(--spacing-lg);
  display: flex;
  gap: var(--spacing-md);
}

.action-btn {
  flex: 1;
  background: var(--color-lime);
  color: #000;
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.action-btn:hover {
  opacity: 0.9;
}

.action-btn.secondary {
  background: var(--color-border);
  color: var(--color-text);
}

.meetings-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.meeting-item {
  display: grid;
  grid-template-columns: 100px 1fr 150px 120px;
  gap: var(--spacing-lg);
  align-items: center;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  transition: all 0.2s ease;
}

.meeting-item:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.meeting-item.past {
  opacity: 0.7;
}

.meeting-time {
  display: flex;
  align-items: center;
}

.time-box {
  background: rgba(184, 243, 107, 0.1);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  text-align: center;
}

.time-hour {
  display: block;
  font-weight: 600;
  color: var(--color-lime);
  font-size: 14px;
}

.time-date {
  display: block;
  font-size: 11px;
  color: var(--color-text-secondary);
}

.meeting-content h3 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: 15px;
  font-weight: 600;
}

.meeting-content p {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.meeting-description {
  font-size: 13px;
  color: var(--color-text);
}

.meeting-badges {
  display: flex;
  gap: var(--spacing-sm);
}

.badge {
  font-size: 11px;
  padding: 4px 8px;
  background: rgba(124, 196, 255, 0.2);
  color: var(--color-sky);
  border-radius: var(--radius-xs);
}

.btn-join {
  background: var(--color-lime);
  color: #000;
  border: none;
  padding: var(--spacing-sm) var(--spacing-lg);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.btn-join:hover {
  opacity: 0.9;
}

.btn-recording,
.btn-details {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  background: var(--color-border);
  color: var(--color-text);
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.btn-recording:hover,
.btn-details:hover {
  background: var(--color-lime);
  color: #000;
}

.no-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-lg);
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.no-content :deep(svg) {
  opacity: 0.5;
}

@media (max-width: 1024px) {
  .meeting-item {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }

  .meetings-grid {
    grid-template-columns: 1fr;
  }
}
</style>
