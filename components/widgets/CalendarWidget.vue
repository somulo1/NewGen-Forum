<template>
  <div class="calendar-widget">
    <div class="widget-header">
      <h3>Upcoming Events</h3>
      <NuxtLink to="/admin/calendar" class="view-all">View all →</NuxtLink>
    </div>
    <div class="events-list">
      <div v-for="event in events" :key="event.id" class="event-item">
        <div class="event-date">
          <span class="date-day">{{ event.dayOfMonth }}</span>
          <span class="date-month">{{ event.month }}</span>
        </div>
        <div class="event-info">
          <h4>{{ event.title }}</h4>
          <p>{{ event.time }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const events = [
  { id: 1, title: 'Team Standup', time: '10:00 AM', dayOfMonth: '21', month: 'Dec' },
  { id: 2, title: 'Client Meeting', time: '2:00 PM', dayOfMonth: '21', month: 'Dec' },
  { id: 3, title: 'Project Review', time: '4:30 PM', dayOfMonth: '22', month: 'Dec' },
]
</script>

<style scoped>
.calendar-widget {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.widget-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.view-all {
  color: var(--color-lime);
  text-decoration: none;
  font-size: 12px;
  font-weight: 600;
}

.events-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.event-item {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.event-item:last-child {
  border-bottom: none;
}

.event-date {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: rgba(184, 243, 107, 0.1);
  border-radius: var(--radius-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  min-width: 50px;
}

.date-day {
  font-weight: 700;
  font-size: 16px;
  color: var(--color-lime);
}

.date-month {
  font-size: 11px;
  color: var(--color-text-secondary);
}

.event-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.event-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>
