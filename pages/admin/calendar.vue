<template>
  <div class="calendar-container">
    <div class="page-header">
      <h1>Calendar</h1>
      <div class="calendar-controls">
        <button class="nav-btn" @click="previousMonth">‹</button>
        <span class="month-year">October 2026</span>
        <button class="nav-btn" @click="nextMonth">›</button>
        <button class="btn-today">Today</button>
        <button class="btn-new">+ New</button>
      </div>
    </div>

    <div class="calendar-filters">
      <button
        v-for="filter in filters"
        :key="filter.id"
        :class="['filter-tag', { active: filter.active }]"
        @click="toggleFilter(filter.id)"
      >
        <span class="filter-dot" :style="{ background: filter.color }"></span>
        {{ filter.label }}
      </button>
    </div>

    <div class="calendar-main">
      <div class="calendar-grid">
        <div class="weekdays">
          <div class="weekday">MON</div>
          <div class="weekday">TUE</div>
          <div class="weekday">WED</div>
          <div class="weekday">THU</div>
          <div class="weekday">FRI</div>
          <div class="weekday">SAT</div>
          <div class="weekday">SUN</div>
        </div>

        <div class="days-grid">
          <div
            v-for="date in calendarDays"
            :key="date.day"
            :class="['day-cell', { 'other-month': date.otherMonth, 'is-today': date.isToday }]"
          >
            <div class="day-number">{{ date.day }}</div>
            <div class="events-in-day">
              <div v-for="event in date.events.slice(0, 2)" :key="event.id" :class="['event-tag', `event-${event.type}`]">
                {{ event.time }} {{ event.title }}
              </div>
              <div v-if="date.events.length > 2" class="more-events">+{{ date.events.length - 2 }} more</div>
            </div>
          </div>
        </div>
      </div>

      <aside class="calendar-sidebar">
        <div class="sidebar-header">
          <h3>Today · Thu, 15 Oct</h3>
        </div>

        <div class="events-sidebar-list">
          <div v-for="event in todayEvents" :key="event.id" class="event-card">
            <div class="event-time-range">{{ event.timeRange }}</div>
            <div :class="['event-badge', `badge-${event.badgeType}`]">{{ event.badgeLabel }}</div>
            <h4>{{ event.title }}</h4>
            <div v-if="event.location" class="event-location">{{ event.location }}</div>
            <div class="event-attendees">
              <div v-for="att in event.attendees" :key="att.initials" class="attendee-avatar" :style="{ background: att.color }">
                {{ att.initials }}
              </div>
              <span class="attendee-count">{{ event.attendees.length }} people</span>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const filters = [
  { id: 1, label: 'My own', color: '#F49AC1', active: true },
  { id: 2, label: 'Meetings', color: '#7CC4FF', active: true },
  { id: 3, label: 'Standups', color: '#6EE7C8', active: true },
  { id: 4, label: 'Tasks', color: '#FFD66B', active: true },
  { id: 5, label: 'Events', color: '#9DB0A6', active: true },
  { id: 6, label: 'Project deadlines', color: '#FFB86B', active: true },
]

const toggleFilter = (id: number) => {
  const filter = filters.find(f => f.id === id)
  if (filter) filter.active = !filter.active
}

const previousMonth = () => {}
const nextMonth = () => {}

const calendarDays = [
  { day: 28, otherMonth: true, isToday: false, events: [] },
  { day: 29, otherMonth: true, isToday: false, events: [] },
  { day: 30, otherMonth: true, isToday: false, events: [] },
  { day: 1, otherMonth: false, isToday: false, events: [] },
  { day: 2, otherMonth: false, isToday: false, events: [] },
  { day: 3, otherMonth: false, isToday: false, events: [] },
  { day: 4, otherMonth: false, isToday: false, events: [] },
  { day: 5, otherMonth: false, isToday: false, events: [] },
  { day: 6, otherMonth: false, isToday: false, events: [] },
  { day: 7, otherMonth: false, isToday: false, events: [] },
  { day: 8, otherMonth: false, isToday: false, events: [] },
  { day: 9, otherMonth: false, isToday: false, events: [] },
  { day: 10, otherMonth: false, isToday: false, events: [
    { id: 1, time: 'Install North...', type: 'deadline', title: 'Install North...' }
  ]},
  { day: 11, otherMonth: false, isToday: false, events: [
    { id: 2, time: '09:15–09:30', type: 'standup', title: 'IT daily standup' },
    { id: 3, time: 'Day-one sche...', type: 'deadline', title: 'Day-one sche...' },
    { id: 4, time: 'Offline cache...', type: 'event', title: 'Offline cache...' }
  ]},
  { day: 12, otherMonth: false, isToday: false, events: [
    { id: 5, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' },
    { id: 6, time: '08:30 Risk & ...', type: 'standup', title: 'Risk & ...' }
  ]},
  { day: 13, otherMonth: false, isToday: false, events: [
    { id: 7, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 14, otherMonth: false, isToday: false, events: [
    { id: 8, time: 'Draft the Q4...', type: 'task', title: 'Draft the Q4...' },
    { id: 9, time: 'Push notificat...', type: 'task', title: 'Push notificat...' }
  ]},
  { day: 15, otherMonth: false, isToday: true, events: [
    { id: 10, time: '09:15', type: 'standup', title: 'IT daily standup' },
    { id: 11, time: '15:00 Sprint...', type: 'standup', title: 'Sprint...' }
  ]},
  { day: 16, otherMonth: false, isToday: false, events: [
    { id: 12, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' },
    { id: 13, time: '17:00 Submit...', type: 'deadline', title: 'Submit...' }
  ]},
  { day: 17, otherMonth: false, isToday: false, events: [
    { id: 14, time: '11:00 Access...', type: 'deadline', title: 'Access...' }
  ]},
  { day: 18, otherMonth: false, isToday: false, events: [
    { id: 15, time: '08:00 Team b...', type: 'standup', title: 'Team b...' }
  ]},
  { day: 19, otherMonth: false, isToday: false, events: [
    { id: 16, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 20, otherMonth: false, isToday: false, events: [
    { id: 17, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' },
    { id: 18, time: 'Draft the Q4...', type: 'task', title: 'Draft the Q4...' }
  ]},
  { day: 21, otherMonth: false, isToday: false, events: [
    { id: 19, time: '08:30 Risk & ...', type: 'standup', title: 'Risk & ...' },
    { id: 20, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 22, otherMonth: false, isToday: false, events: [
    { id: 21, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 23, otherMonth: false, isToday: false, events: [
    { id: 22, time: '10:00 Quarter...', type: 'meeting', title: 'Quarter...' },
    { id: 23, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 24, otherMonth: false, isToday: false, events: [
    { id: 24, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 25, otherMonth: false, isToday: false, events: [
    { id: 25, time: 'Order Coast b...', type: 'task', title: 'Order Coast b...' }
  ]},
  { day: 26, otherMonth: false, isToday: false, events: [
    { id: 26, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 27, otherMonth: false, isToday: false, events: [
    { id: 27, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 28, otherMonth: false, isToday: false, events: [
    { id: 28, time: '08:30 Risk & ...', type: 'standup', title: 'Risk & ...' },
    { id: 29, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 29, otherMonth: false, isToday: false, events: [
    { id: 30, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 30, otherMonth: false, isToday: false, events: [
    { id: 31, time: '09:15 IT daily...', type: 'standup', title: 'IT daily...' }
  ]},
  { day: 31, otherMonth: false, isToday: false, events: [
    { id: 32, time: 'Test failover li...', type: 'deadline', title: 'Test failover li...' }
  ]},
  { day: 1, otherMonth: true, isToday: false, events: [] },
]

const todayEvents = [
  {
    id: 1,
    timeRange: '09:15–09:30',
    badgeType: 'standup',
    badgeLabel: 'Standup',
    title: 'IT daily standup',
    location: '',
    attendees: [
      { initials: 'BK', color: '#7CC4FF' },
      { initials: 'DO', color: '#F49AC1' },
      { initials: 'GN', color: '#FFB86B' }
    ]
  },
  {
    id: 2,
    timeRange: '15:00–16:00',
    badgeType: 'meeting',
    badgeLabel: 'Meeting',
    title: 'Sprint review',
    location: 'Conference room B',
    attendees: [
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'FA', color: '#B8F36B' },
      { initials: 'BK', color: '#7CC4FF' }
    ]
  }
]
</script>

<style scoped>
.calendar-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
  color: var(--color-text);
}

.calendar-controls {
  display: flex;
  align-items: center;
  gap: 16px;
}

.nav-btn {
  background: transparent;
  border: none;
  color: var(--color-text);
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
}

.month-year {
  font-weight: 600;
  min-width: 120px;
  text-align: center;
}

.btn-today {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}

.btn-new {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}

.calendar-filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  color: var(--color-text);
  transition: all 0.2s ease;
}

.filter-tag.active {
  background: var(--color-border);
  border-color: var(--color-border);
}

.filter-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.calendar-main {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 24px;
}

.calendar-grid {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8px;
  margin-bottom: 12px;
}

.weekday {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  padding: 8px 0;
  text-align: center;
}

.days-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8px;
  min-height: 400px;
}

.day-cell {
  aspect-ratio: 1;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative;
  background: var(--color-bg);
}

.day-cell.other-month {
  opacity: 0.3;
}

.day-cell.is-today {
  background: var(--color-surface);
  border: 2px solid #B8F36B;
}

.day-number {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text);
}

.day-cell.is-today .day-number {
  width: 20px;
  height: 20px;
  background: #B8F36B;
  color: #000;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
}

.events-in-day {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.event-tag {
  font-size: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

.event-standup {
  color: #6EE7C8;
}

.event-deadline {
  color: #FF9B8A;
}

.event-task {
  color: #FFD66B;
}

.event-meeting {
  color: #7CC4FF;
}

.event-event {
  color: #9B8FFF;
}

.more-events {
  font-size: 10px;
  color: var(--color-text-secondary);
}

.calendar-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.events-sidebar-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.event-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 12px;
}

.event-time-range {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}

.event-badge {
  display: inline-block;
  font-size: 10px;
  padding: 3px 6px;
  border-radius: 3px;
  font-weight: 600;
  margin-bottom: 4px;
}

.badge-standup {
  background: rgba(110, 231, 200, 0.2);
  color: #6EE7C8;
}

.badge-meeting {
  background: rgba(124, 196, 255, 0.2);
  color: #7CC4FF;
}

.badge-deadline {
  background: rgba(255, 155, 138, 0.2);
  color: #FF9B8A;
}

.event-card h4 {
  margin: 4px 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.event-location {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin: 4px 0;
}

.event-attendees {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 8px 0;
  flex-wrap: wrap;
}

.attendee-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 700;
  color: white;
}

.attendee-count {
  font-size: 11px;
  color: var(--color-text-secondary);
}

@media (max-width: 1024px) {
  .calendar-main {
    grid-template-columns: 1fr;
  }

  .calendar-sidebar {
    order: -1;
  }
}
</style>
