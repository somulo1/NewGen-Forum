<template>
  <div class="sessions-container">
    <div class="page-header">
      <div>
        <h1>Meetings & standups</h1>
        <p class="subtitle">Schedule meetings and standups, pick who joins, and start them from here.</p>
      </div>
      <div class="header-buttons">
        <button class="btn-schedule">Schedule a meeting</button>
        <button class="btn-standup">Schedule a standup</button>
        <button class="btn-start">Start now</button>
      </div>
    </div>

    <div class="meetings-table">
      <div class="table-header">
        <div class="col-meeting">MEETING</div>
        <div class="col-type">TYPE</div>
        <div class="col-date">DATE</div>
        <div class="col-time">TIME</div>
        <div class="col-repeats">REPEATS</div>
        <div class="col-participants">PARTICIPANTS</div>
      </div>

      <div class="table-filters">
        <input type="text" class="filter-input" placeholder="Filter...">
        <select class="filter-select">
          <option>All</option>
        </select>
        <input type="date" class="filter-input">
        <input type="date" class="filter-input">
        <select class="filter-select">
          <option>All</option>
        </select>
      </div>

      <div v-for="meeting in meetings" :key="meeting.id" class="table-row">
        <div class="col-meeting">
          <div class="meeting-info">
            <div class="meeting-title">{{ meeting.title }}</div>
            <div class="meeting-dept">{{ meeting.department }}</div>
            <div class="meeting-forum">{{ meeting.forum }}</div>
          </div>
        </div>
        <div class="col-type">
          <span class="type-badge" :class="'type-' + meeting.type.toLowerCase()">{{ meeting.type }}</span>
        </div>
        <div class="col-date">{{ meeting.date }}</div>
        <div class="col-time">{{ meeting.time }}</div>
        <div class="col-repeats">{{ meeting.repeats }}</div>
        <div class="col-participants">
          <div class="participants-avatars">
            <div v-for="(person, idx) in meeting.participants" :key="idx" class="participant-avatar" :style="{ background: person.color }">
              {{ person.initials }}
            </div>
            <span class="participant-count">{{ meeting.participants.length }} people</span>
          </div>
        </div>
      </div>
    </div>

    <div class="pagination">
      <button class="page-btn">Previous</button>
      <span class="page-info">Showing 1–5 of 5</span>
      <button class="page-btn">Next</button>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})

const meetings = [
  {
    id: 1,
    title: 'IT daily standup',
    department: 'IT Operations',
    forum: 'forum',
    type: 'Standup',
    date: 'Tue, 13 Oct',
    time: '09:15–09:30',
    repeats: 'Weekdays',
    participants: [
      { initials: 'BK', color: '#7CC4FF' },
      { initials: 'DO', color: '#F49AC1' },
      { initials: 'JK', color: '#FFB86B' },
    ],
  },
  {
    id: 2,
    title: 'Risk & Compliance weekly sync',
    department: 'Risk & Compliance',
    forum: 'forum',
    type: 'Standup',
    date: 'Wed, 14 Oct',
    time: '08:30–08:45',
    repeats: 'Weekly',
    participants: [
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'KM', color: '#9B8FFF' },
    ],
  },
  {
    id: 3,
    title: 'Sprint review',
    department: 'Product Builders',
    forum: 'forum',
    type: 'Meeting',
    date: 'Thu, 15 Oct',
    time: '15:00–16:00',
    repeats: 'Once',
    participants: [
      { initials: 'DO', color: '#F49AC1' },
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'BK', color: '#7CC4FF' },
      { initials: 'KM', color: '#9B8FFF' },
    ],
  },
  {
    id: 4,
    title: 'Access review with department heads',
    department: 'Risk & Compliance',
    forum: 'forum',
    type: 'Meeting',
    date: 'Sat, 17 Oct',
    time: '11:00–12:00',
    repeats: 'Once',
    participants: [
      { initials: 'GN', color: '#FFB86B' },
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'KM', color: '#9B8FFF' },
    ],
  },
  {
    id: 5,
    title: 'Quarterly town hall',
    department: 'Finance',
    forum: 'forum',
    type: 'Meeting',
    date: 'Thu, 22 Oct',
    time: '10:00–11:30',
    repeats: 'Once',
    participants: [
      { initials: 'BK', color: '#7CC4FF' },
      { initials: 'GN', color: '#FFB86B' },
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'FA', color: '#B8F36B' },
      { initials: 'KM', color: '#9B8FFF' },
      { initials: 'JK', color: '#FFB86B' },
    ],
  },
]
</script>

<style scoped>
.sessions-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: 8px 0 0 0;
}

.header-buttons {
  display: flex;
  gap: 12px;
}

.btn-schedule,
.btn-standup {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn-schedule:hover,
.btn-standup:hover {
  border-color: #B8F36B;
  color: #B8F36B;
}

.btn-start {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.meetings-table {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 200px 100px 120px 120px 100px 200px;
  gap: 0;
  background: var(--color-border);
  padding: 12px 16px;
  font-weight: 600;
  font-size: 11px;
  color: var(--color-text-secondary);
  text-transform: uppercase;
}

.table-filters {
  display: grid;
  grid-template-columns: 200px 100px 120px 120px 100px 200px;
  gap: 0;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
}

.filter-input,
.filter-select {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 6px 8px;
  font-size: 12px;
  color: var(--color-text);
  cursor: pointer;
}

.table-row {
  display: grid;
  grid-template-columns: 200px 100px 120px 120px 100px 200px;
  gap: 0;
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
}

.table-row:hover {
  background: var(--color-bg);
}

.meeting-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meeting-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.meeting-dept,
.meeting-forum {
  font-size: 11px;
  color: var(--color-text-secondary);
}

.type-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.type-standup {
  background: rgba(110, 231, 200, 0.2);
  color: #6EE7C8;
}

.type-meeting {
  background: rgba(124, 196, 255, 0.2);
  color: #7CC4FF;
}

.col-date,
.col-time,
.col-repeats {
  font-size: 13px;
  color: var(--color-text);
}

.participants-avatars {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.participant-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 700;
  color: white;
}

.participant-count {
  font-size: 11px;
  color: var(--color-text-secondary);
  margin-left: 4px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
}

.page-btn {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.page-btn:hover {
  border-color: #B8F36B;
  color: var(--color-text);
}

.page-info {
  font-size: 12px;
  color: var(--color-text-secondary);
}

@media (max-width: 1024px) {
  .table-header,
  .table-filters,
  .table-row {
    grid-template-columns: 1fr;
  }
  
  .header-buttons {
    flex-wrap: wrap;
  }
}
</style>
