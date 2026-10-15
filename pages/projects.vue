<template>
  <div class="projects-container">
    <div class="page-header">
      <div>
        <h1>Projects</h1>
        <p class="subtitle">Every project's goal, progress, milestones and risks in one place.</p>
      </div>
      <button class="btn-new-project">
        New project
      </button>
    </div>

    <div class="project-stats">
      <div class="stat-card">
        <div class="stat-number">4</div>
        <div class="stat-label">Active projects</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">2</div>
        <div class="stat-label">On track</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">1</div>
        <div class="stat-label">At risk</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">1</div>
        <div class="stat-label">Off track</div>
      </div>
    </div>

    <div class="integration-bar">
      <div class="integration-item">
        <span class="integration-name">Slack · connected</span>
      </div>
      <div class="integration-item">
        <span class="integration-name">GitHub · connected</span>
      </div>
      <div class="integration-item">
        <span class="integration-name">Trello · connected</span>
      </div>
      <div class="integration-item">
        <span class="integration-name">+ Connect Asana, ClickUp...</span>
      </div>
    </div>

    <div class="projects-grid">
      <div v-for="project in projects" :key="project.id" class="project-card">
        <div class="card-content">
          <div class="project-avatar" :style="{ background: project.color }">{{ project.initials }}</div>
          <div class="project-info">
            <div class="project-title-row">
              <h3>{{ project.name }}</h3>
              <span class="status-badge" :class="'status-' + project.status.toLowerCase().replace(' ', '-')">{{ project.status }}</span>
            </div>
            <p class="project-id">{{ project.id }} · Led by {{ project.lead }}</p>
            <p class="project-description">{{ project.description }}</p>

            <div class="progress-section">
              <div class="progress-label">Progress</div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: project.progress + '%', background: project.progressColor }"></div>
              </div>
              <div class="progress-value">{{ project.progress }}%</div>
            </div>

            <div class="milestone-section">
              <div class="milestone-row">
                <span class="milestone-label">Next milestone</span>
                <span class="milestone-label">Due</span>
                <span class="milestone-label">Open risks</span>
              </div>
              <div class="milestone-row values">
                <span class="milestone-content">{{ project.nextMilestone }}</span>
                <span class="milestone-content">{{ project.dueDate }}</span>
                <span class="milestone-content">{{ project.openRisks }}</span>
              </div>
            </div>

            <div class="card-footer">
              <div class="team-avatars">
                <div v-for="(member, idx) in project.members" :key="idx" class="member-avatar" :style="{ background: member.color }">{{ member.initials }}</div>
                <span class="member-count">{{ project.members.length }} people · {{ project.team }}</span>
              </div>
              <button class="btn-view">View</button>
            </div>
          </div>
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

const projects = [
  {
    id: 'PRJ-04',
    name: 'Mobile App 2.0',
    initials: 'MA',
    color: '#7CC4FF',
    lead: 'David Otieno',
    description: 'Give every staff member a fast mobile app with offline mode and a dark theme, so field teams can keep working without signal.',
    status: 'On track',
    progress: 38,
    progressColor: '#B8F36B',
    nextMilestone: 'Offline mode',
    dueDate: '15 Dec 2026',
    openRisks: '2',
    members: [
      { initials: 'DO', color: '#F49AC1' },
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'BK', color: '#7CC4FF' },
      { initials: 'KM', color: '#9B8FFF' }
    ],
    team: 'mobile-app'
  },
  {
    id: 'PRJ-02',
    name: 'Branch Network Upgrade',
    initials: 'BN',
    color: '#6EE7C8',
    lead: 'Brian Kiprop',
    description: 'Replace old routers and Wi-Fi at every branch so staff get stable connections and fewer network tickets.',
    status: 'At risk',
    progress: 33,
    progressColor: '#FFD66B',
    nextMilestone: 'North branch',
    dueDate: '30 Nov 2026',
    openRisks: '2',
    members: [
      { initials: 'BK', color: '#7CC4FF' },
      { initials: 'DO', color: '#F49AC1' },
      { initials: 'FA', color: '#B8F36B' },
      { initials: 'JK', color: '#FFB86B' }
    ],
    team: 'network-upgrade'
  },
  {
    id: 'PRJ-05',
    name: 'Onboarding Revamp',
    initials: 'OR',
    color: '#F49AC1',
    lead: 'Grace Njeri',
    description: 'Get new staff fully set up on their first day instead of their third.',
    status: 'On track',
    progress: 20,
    progressColor: '#B8F36B',
    nextMilestone: 'Account automation',
    dueDate: '20 Nov 2026',
    openRisks: '1',
    members: [
      { initials: 'GN', color: '#FFB86B' },
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'BK', color: '#7CC4FF' }
    ],
    team: 'onboarding'
  },
  {
    id: 'PRJ-06',
    name: 'Year-end Audit Prep',
    initials: 'YA',
    color: '#FFB86B',
    lead: 'Amina Wanjiru',
    description: 'Have every control, policy and evidence file ready before the external auditors arrive.',
    status: 'Off track',
    progress: 20,
    progressColor: '#FF9B8A',
    nextMilestone: 'Control testing',
    dueDate: '18 Dec 2026',
    openRisks: '2',
    members: [
      { initials: 'AW', color: '#6EE7C8' },
      { initials: 'FA', color: '#B8F36B' },
      { initials: 'GN', color: '#FFB86B' }
    ],
    team: 'audit-prep'
  }
]
</script>

<style scoped>
.projects-container {
  display: flex;
  flex-direction: column;
  gap: 32px;
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
  color: var(--color-text);
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: 8px 0 0 0;
}

.btn-new-project {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.btn-new-project:hover {
  opacity: 0.9;
}

.project-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px 16px;
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  color: #B8F36B;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.integration-bar {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
}

.integration-item {
  padding: 8px 12px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.integration-name {
  white-space: nowrap;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.project-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  padding: 0;
  overflow: hidden;
}

.card-content {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 24px;
  padding: 24px;
}

.project-avatar {
  width: 120px;
  height: 120px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
}

.project-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.project-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.project-title-row h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.status-badge {
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 600;
  white-space: nowrap;
}

.status-on-track {
  background: rgba(184, 243, 107, 0.2);
  color: #B8F36B;
}

.status-at-risk {
  background: rgba(255, 214, 107, 0.2);
  color: #FFD66B;
}

.status-off-track {
  background: rgba(255, 155, 138, 0.2);
  color: #FF9B8A;
}

.project-id {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.project-description {
  margin: 4px 0 0 0;
  font-size: 13px;
  color: var(--color-text);
  line-height: 1.5;
}

.progress-section {
  margin-top: 8px;
}

.progress-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-bottom: 6px;
}

.progress-bar {
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 3px;
}

.progress-value {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text);
  margin-top: 4px;
  text-align: right;
}

.milestone-section {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--color-border);
}

.milestone-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 16px;
  margin-bottom: 4px;
}

.milestone-row.values {
  margin-bottom: 0;
}

.milestone-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: capitalize;
}

.milestone-content {
  font-size: 12px;
  color: var(--color-text);
  font-weight: 500;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12px;
}

.team-avatars {
  display: flex;
  align-items: center;
  gap: 6px;
}

.member-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  color: white;
}

.member-count {
  font-size: 11px;
  color: var(--color-text-secondary);
  margin-left: 4px;
}

.btn-view {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.btn-view:hover {
  opacity: 0.9;
}

@media (max-width: 1024px) {
  .projects-grid {
    grid-template-columns: 1fr;
  }

  .project-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .integration-bar {
    flex-wrap: wrap;
  }
}
</style>
