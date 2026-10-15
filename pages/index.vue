<template>
  <div class="home-container">
    <div class="home-header">
      <div>
        <h1>Good afternoon, Sam</h1>
        <p class="subtitle">Your home page. Add, remove and arrange widgets to suit your organization.</p>
      </div>
      <div class="header-buttons">
        <button class="btn-customize">Customize</button>
        <button class="btn-add-widget" @click="showAddWidget = true">
          <Icon name="plus" size="16" />
          Add widget
        </button>
      </div>
    </div>

    <div class="widgets-grid">
      <!-- Performance Awards -->
      <div class="widget-card performance-awards">
        <div class="widget-section">
          <div class="widget-title-group">
            <Icon name="award" size="20" class="title-icon" />
            <div>
              <h3>Performance awards</h3>
              <p>10 people recognised recently · tap anyone to see their profile</p>
            </div>
          </div>
          <button class="btn-award">Give an award</button>
        </div>
        <div class="awards-grid">
          <div v-for="award in performanceAwards" :key="award.id" class="award-card">
            <div class="award-avatar" :style="{ background: award.color }">{{ award.initials }}</div>
            <h4>{{ award.name }}</h4>
            <p class="award-title">{{ award.achievement }}</p>
            <p class="award-date">{{ award.date }}</p>
          </div>
        </div>
      </div>

      <!-- Announcements -->
      <div class="widget-card announcements">
        <div class="widget-header-section">
          <div class="status-indicator green"></div>
          <h3>Announcements</h3>
          <p>From management</p>
        </div>
        <div class="announcements-list">
          <div v-for="announcement in announcements" :key="announcement.id" class="announcement-row">
            <span class="tag" :class="`tag-${announcement.tagType}`">{{ announcement.tag }}</span>
            <div class="announcement-text">
              <h4>{{ announcement.title }}</h4>
              <p>{{ announcement.author }}</p>
            </div>
            <button class="btn-mark">Mark read</button>
          </div>
        </div>
        <div class="widget-footer">
          <button class="btn-new">New announcement</button>
          <a href="#" class="link-all">See all</a>
        </div>
      </div>

      <!-- Meetings & Standups -->
      <div class="widget-card meetings">
        <div class="widget-header-section">
          <div class="status-indicator green"></div>
          <h3>Meetings & standups</h3>
          <p>Live & upcoming</p>
        </div>
        <div class="meetings-list">
          <div v-for="meeting in meetings" :key="meeting.id" class="meeting-row">
            <div class="meeting-badge" :class="`badge-${meeting.badgeClass}`">{{ meeting.badge }}</div>
            <div class="meeting-text">
              <h4>{{ meeting.title }}</h4>
              <p>{{ meeting.time }}</p>
            </div>
            <span v-if="meeting.live" class="live-label">LIVE</span>
          </div>
        </div>
        <div class="widget-actions">
          <button class="btn-start-meeting">Start a meeting</button>
          <button class="btn-schedule">Schedule</button>
        </div>
      </div>

      <!-- Departments -->
      <div class="widget-card departments">
        <h3>Departments</h3>
        <div class="departments-list">
          <div v-for="dept in departments" :key="dept.id" class="department-row">
            <div class="dept-info">
              <h4>{{ dept.name }}</h4>
              <p>{{ dept.count }} people</p>
            </div>
            <div class="dept-stats">
              <span class="percentage">{{ dept.percentage }}%</span>
              <a href="#" class="view-link">View</a>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Widget Modal -->
    <Teleport to="body">
      <div v-if="showAddWidget" class="modal-overlay" @click="showAddWidget = false">
        <div class="modal-container" @click.stop>
          <div class="modal-header">
            <div>
              <h2>Add a widget</h2>
              <p>Pick what your home page should show. You can remove or reorder them anytime.</p>
            </div>
            <button class="modal-close" @click="showAddWidget = false">
              <Icon name="x" size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div class="widgets-options-grid">
              <div
                v-for="widget in availableWidgets"
                :key="widget.id"
                class="widget-option"
                :class="{ 'is-added': widget.added }"
              >
                <div class="option-top">
                  <div class="dot-indicator" :style="{ background: widget.color }"></div>
                  <h4>{{ widget.name }}</h4>
                </div>
                <p class="option-description">{{ widget.description }}</p>
                <div class="option-bottom">
                  <button
                    class="btn-toggle"
                    :class="{ added: widget.added }"
                  >
                    {{ widget.added ? 'Added' : 'Add to home' }}
                  </button>
                  <Icon name="settings" size="16" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const showAddWidget = ref(false)

const performanceAwards = [
  {
    id: 1,
    name: 'David Otieno',
    initials: 'DO',
    achievement: 'Star performer · Q3',
    date: 'Mon 28 Sept',
    color: 'linear-gradient(135deg, #F49AC1 0%, #C9B6FF 100%)',
  },
  {
    id: 2,
    name: 'Grace Njeri',
    initials: 'GN',
    achievement: 'Star performer · Q3',
    date: 'Mon 28 Sept',
    color: 'linear-gradient(135deg, #FFB86B 0%, #FFD66B 100%)',
  },
  {
    id: 3,
    name: 'Kevin Mutua',
    initials: 'KM',
    achievement: 'Innovation award',
    date: 'Tue 15 Sept',
    color: 'linear-gradient(135deg, #C9B6FF 0%, #7CC4FF 100%)',
  },
  {
    id: 4,
    name: 'John Kamau',
    initials: 'JK',
    achievement: 'Team player',
    date: 'Tue 1 Sept',
    color: 'linear-gradient(135deg, #7CC4FF 0%, #6EE7C8 100%)',
  },
]

const announcements = [
  {
    id: 1,
    tag: 'Policy',
    tagType: 'green',
    title: 'Updated leave policy takes effect 1 November',
    author: 'HR · 2 days ago',
  },
  {
    id: 2,
    tag: 'Office',
    tagType: 'orange',
    title: 'Office closed Friday for maintenance',
    author: 'Admin · Today',
  },
  {
    id: 3,
    tag: 'IT',
    tagType: 'green',
    title: 'Planned downtime Saturday 8 to 10pm',
    author: 'IT Operations · Yesterday',
  },
]

const meetings = [
  {
    id: 1,
    badge: 'IT',
    badgeClass: 'green',
    title: 'IT daily standup',
    time: 'Weekdays 9:15',
  },
  {
    id: 2,
    badge: 'PB',
    badgeClass: 'blue',
    title: 'Sprint review',
    time: 'Now',
    live: true,
  },
  {
    id: 3,
    badge: 'CW',
    badgeClass: 'pink',
    title: 'Choir rehearsal (online)',
    time: 'Thu 5:30pm',
  },
  {
    id: 4,
    badge: 'IT',
    badgeClass: 'green',
    title: 'IT town hall',
    time: 'Fri 10:00',
  },
]

const departments = [
  { id: 1, name: 'Development', count: 23, percentage: 88 },
  { id: 2, name: 'Design', count: 8, percentage: 96 },
  { id: 3, name: 'Operations', count: 15, percentage: 95 },
  { id: 4, name: 'Marketing', count: 12, percentage: 91 },
]

const availableWidgets = [
  {
    id: 'world-clock',
    name: 'World clock',
    description: 'Live time in the cities your teams work from',
    color: '#6EE7C8',
    added: false,
  },
  {
    id: 'celebrations',
    name: 'Celebrations',
    description: 'Who is celebrating today, with a quick way to send wishes',
    color: '#F49AC1',
    added: false,
  },
  {
    id: 'tickets',
    name: 'Tickets',
    description: 'Open, overdue and due-soon tickets. The most urgent ones to view, and a quick way to raise a new one.',
    color: '#FFB86B',
    added: true,
  },
  {
    id: 'projects',
    name: 'Projects',
    description: 'Every project with its status, progress and due date, plus a View button into each one.',
    color: '#6EE7C8',
    added: true,
  },
  {
    id: 'performance-awards',
    name: 'Performance awards',
    description: 'Spotlight top performers, show award winners and let staff nominate colleagues.',
    color: '#FFD66B',
    added: true,
  },
  {
    id: 'secure-share',
    name: 'Secure share',
    description: 'Share a password or credential through a link that self-destructs after one view.',
    color: '#FFB86B',
    added: false,
  },
  {
    id: 'community-wall',
    name: 'Community wall',
    description: 'The newest posts from the wall, so nobody misses what colleagues are sharing.',
    color: '#6EE7C8',
    added: false,
  },
  {
    id: 'my-forums',
    name: 'My forums',
    description: 'Your forums with unread counts. Open any one straight into the side panel.',
    color: '#7CC4FF',
    added: false,
  },
  {
    id: 'people-finder',
    name: 'People finder',
    description: 'Search anyone in the organization and see who is online.',
    color: '#7CC4FF',
    added: false,
  },
]
</script>

<style scoped>
.home-container {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.home-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.home-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  color: var(--color-text);
}

.home-header .subtitle {
  margin: 8px 0 0 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.header-buttons {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.btn-customize {
  padding: 8px 20px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 20px;
  color: var(--color-text);
  cursor: pointer;
  font-weight: 500;
  font-size: 13px;
  transition: all 0.2s ease;
}

.btn-customize:hover {
  border-color: var(--color-lime);
  color: var(--color-lime);
}

.btn-add-widget {
  padding: 8px 20px;
  background: var(--color-lime);
  border: none;
  border-radius: 20px;
  color: #000;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: opacity 0.2s ease;
}

.btn-add-widget:hover {
  opacity: 0.9;
}

.widgets-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.widget-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  padding: 24px;
}

.performance-awards {
  grid-column: 1 / -1;
}

.widget-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.widget-title-group {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.title-icon {
  color: var(--color-text);
  margin-top: 2px;
}

.widget-title-group h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.widget-title-group p {
  margin: 4px 0 0 0;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.btn-award {
  padding: 8px 16px;
  background: #FFD66B;
  border: none;
  border-radius: 8px;
  color: #000;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.btn-award:hover {
  opacity: 0.9;
}

.awards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.award-card {
  background: rgba(184, 243, 107, 0.08);
  border: 1px solid rgba(184, 243, 107, 0.15);
  border-radius: 12px;
  padding: 16px;
  text-align: center;
}

.award-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  margin: 0 auto 12px;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 18px;
}

.award-card h4 {
  margin: 0 0 4px 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.award-title {
  margin: 4px 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.award-date {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-secondary);
}

.widget-header-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-indicator.green {
  background: var(--color-lime);
}

.widget-header-section h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.widget-header-section p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.announcements-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.announcement-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
}

.announcement-row:last-child {
  border-bottom: none;
}

.tag {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}

.tag-green {
  background: rgba(110, 231, 200, 0.2);
  color: var(--color-teal);
}

.tag-orange {
  background: rgba(255, 182, 107, 0.2);
  color: var(--color-orange);
}

.announcement-text {
  flex: 1;
}

.announcement-text h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.announcement-text p {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.btn-mark {
  padding: 4px 12px;
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
  transition: color 0.2s ease;
}

.btn-mark:hover {
  color: var(--color-lime);
}

.widget-footer {
  display: flex;
  gap: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.btn-new {
  padding: 8px 16px;
  background: var(--color-lime);
  border: none;
  border-radius: 8px;
  color: #000;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.btn-new:hover {
  opacity: 0.9;
}

.link-all {
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: 12px;
  transition: color 0.2s ease;
}

.link-all:hover {
  color: var(--color-lime);
}

.meetings-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.meeting-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
}

.meeting-row:last-child {
  border-bottom: none;
}

.meeting-badge {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 12px;
  color: white;
  flex-shrink: 0;
}

.badge-green {
  background: var(--color-lime);
  color: #000;
}

.badge-blue {
  background: var(--color-sky);
  color: #000;
}

.badge-pink {
  background: var(--color-pink);
  color: white;
}

.meeting-text {
  flex: 1;
}

.meeting-text h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.meeting-text p {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.live-label {
  background: var(--color-coral);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 700;
  white-space: nowrap;
}

.widget-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.btn-start-meeting {
  flex: 1;
  padding: 8px 16px;
  background: var(--color-lime);
  border: none;
  border-radius: 8px;
  color: #000;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.btn-start-meeting:hover {
  opacity: 0.9;
}

.btn-schedule {
  flex: 1;
  padding: 8px 16px;
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text);
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.btn-schedule:hover {
  border-color: var(--color-lime);
  color: var(--color-lime);
}

.departments h3 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.departments-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.department-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
}

.department-row:last-child {
  border-bottom: none;
}

.dept-info h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.dept-info p {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.dept-stats {
  display: flex;
  align-items: center;
  gap: 16px;
}

.percentage {
  color: var(--color-lime);
  font-weight: 600;
  font-size: 13px;
}

.view-link {
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: 12px;
  transition: color 0.2s ease;
}

.view-link:hover {
  color: var(--color-lime);
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  width: 90%;
  max-width: 900px;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 24px;
  border-bottom: 1px solid var(--color-border);
}

.modal-header h2 {
  margin: 0 0 8px 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
}

.modal-header p {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.modal-close {
  background: none;
  border: none;
  color: var(--color-text);
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.modal-body {
  padding: 24px;
}

.widgets-options-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.widget-option {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: all 0.2s ease;
}

.widget-option:hover {
  border-color: var(--color-lime);
  background: rgba(184, 243, 107, 0.05);
}

.widget-option.is-added {
  opacity: 0.6;
}

.option-top {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dot-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.option-top h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.option-description {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
  line-height: 1.5;
  flex: 1;
}

.option-bottom {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: space-between;
}

.btn-toggle {
  flex: 1;
  padding: 6px 12px;
  background: var(--color-lime);
  border: none;
  border-radius: 6px;
  color: #000;
  cursor: pointer;
  font-weight: 600;
  font-size: 12px;
  transition: opacity 0.2s ease;
}

.btn-toggle:hover {
  opacity: 0.9;
}

.btn-toggle.added {
  background: var(--color-border);
  color: var(--color-text-secondary);
  cursor: not-allowed;
}

@media (max-width: 1024px) {
  .widgets-grid {
    grid-template-columns: 1fr;
  }

  .awards-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .widgets-options-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .home-header {
    flex-direction: column;
  }

  .header-buttons {
    width: 100%;
  }

  .btn-customize,
  .btn-add-widget {
    flex: 1;
  }

  .awards-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .widgets-options-grid {
    grid-template-columns: 1fr;
  }

  .widget-section {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
